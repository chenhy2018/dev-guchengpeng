package stgapi

import (
	"io"
	"strings"
	"sync"

	"qbox.us/ptfd/stgapi.v1/api"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type reader struct {
	xl      *xlog.Logger
	stg     Stg
	cfg     Cfg
	idc     string
	eblocks []string
	from    int64
	to      int64
	err     error
	rc      io.ReadCloser
}

func newReader(xl *xlog.Logger, stg Stg, cfg Cfg, idc string, eblocks []string, from, to int64) io.ReadCloser {

	return &reader{
		xl:      xl,
		stg:     stg,
		cfg:     cfg,
		idc:     idc,
		eblocks: eblocks,
		from:    from,
		to:      to,
	}
}

func (p *reader) Read(b []byte) (int, error) {

	if p.err != nil {
		return 0, p.err
	}
	if p.from == p.to {
		return 0, io.EOF
	}

	if p.rc == nil {
		ifrom := p.from / api.MaxDataSize
		fromBase := ifrom * api.MaxDataSize
		from := uint32(p.from - fromBase)
		to := uint32(api.MaxDataSize)
		if fromBase+api.MaxDataSize > p.to {
			to = uint32(p.to - fromBase)
		}
		eblock := p.eblocks[ifrom]
		rc, err := newBlockReader(p.xl, p.stg, p.cfg, p.idc, eblock, from, to)
		if err != nil {
			return 0, errors.Info(err, "newBlockReader").Detail(err)
		}
		p.rc = rc
	}

	n, err := p.rc.Read(b)
	p.from += int64(n)

	if p.from == p.to {
		p.err = io.EOF
		return n, p.err
	}

	if err == io.EOF {
		p.rc.Close()
		p.rc = nil
		if n == 0 {
			return p.Read(b)
		}
		err = nil
	}
	return n, err
}

func (p *reader) Close() error {

	if p.rc != nil {
		return p.rc.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------

type blockReader struct {
	xl     *xlog.Logger
	stg    Stg
	eblock string
	from   uint32
	to     uint32
	hosts  []string
	ihost  int
	trys   int
	proxy  bool

	// 所有的(*blockReader).Read()都是顺序执行的,本身是并不需要锁的
	// 但是(*blockReader).Close()是异步操作,有并发问题,大多数情况下,Close只会被调用一次
	// 原来的问题是, 当blockReader被关闭后, p.rc.Read(b)得到了一个被关闭的错误, 但是误以为这是要重试的错误。
	// 就进入重试的逻辑, 更换了一个p.rc。
	// 但是外面的代码已经调用了一次(*blockReader).Close(),认为句柄已经关闭,就造成了句柄泄漏。
	mu     sync.Mutex
	closed bool
	rc     io.ReadCloser
}

func newBlockReader(xl *xlog.Logger, stg Stg, cfg Cfg, idc string, eblock string, from, to uint32) (io.ReadCloser, error) {

	addr, err := api.DecodeEblock(eblock)
	if err != nil {
		return nil, errors.Info(err, "DecodeEblock", eblock).Detail(err)
	}
	hosts, ihost, rIdc, err := cfg.HostsIdc(xl, addr.Dgid)
	if err != nil {
		return nil, errors.Info(ErrServiceUnavailable, "client.DgHosts").Detail(err)
	}
	proxy := false
	if idc != rIdc {
		xl.Infof("newBlockReader: selfIdc(%v) != remoteIdc(%v), use proxy.", idc, rIdc)
		proxy = true
	}
	trys := 0
	nhost := len(hosts)
	var rc io.ReadCloser
	for {
		host := hosts[(ihost+trys)%nhost]
		if !proxy {
			rc, err = stg.Get(xl, host, eblock, from, to)
		} else {
			rc, err = stg.ProxyGet(xl, host, eblock, from, to)
		}
		if err == nil {
			break
		}

		if proxy {
			if e, ok := err.(rpc.RespError); (ok && e.HttpCode() == 503) || (!ok && strings.Contains(err.Error(), "connecting to proxy")) {
				xl.Error("all proxy fails", err)
				return nil, errors.Info(err, "stg.Get", host).Detail(err)
			}
		}
		trys++
		if trys >= nhost {
			return nil, errors.Info(err, "stg.Get", host).Detail(err)
		}
		xl.Errorf("newBlockReader: stg.Get %v failed => %v", host, err)
	}
	return &blockReader{
		xl:     xl,
		stg:    stg,
		eblock: eblock,
		from:   from,
		to:     to,
		hosts:  hosts,
		ihost:  ihost,
		trys:   trys,
		proxy:  proxy,
		rc:     rc,
		closed: false,
	}, nil
}

func (p *blockReader) Read(b []byte) (int, error) {

	p.mu.Lock()
	r := p.rc
	p.mu.Unlock()
	n, err := r.Read(b)
	p.from += uint32(n)
	if p.from == p.to {
		return n, io.EOF
	}
	if err == nil {
		return n, err
	} else {
		p.mu.Lock()
		closed := p.closed
		p.mu.Unlock()
		if closed {
			return 0, err
		}
	}
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}

	nhost := len(p.hosts)
	host := p.hosts[(p.ihost+p.trys)%nhost]
	p.trys++
	if p.trys >= nhost {
		return n, errors.Info(err, "rc.Read", host).Detail(err)
	}
	p.xl.Errorf("blockReader.Read: rc.Read %v failed => %v", host, err)

	for {
		host = p.hosts[(p.ihost+p.trys)%nhost]
		var rc io.ReadCloser
		if !p.proxy {
			rc, err = p.stg.Get(p.xl, host, p.eblock, p.from, p.to)
		} else {
			rc, err = p.stg.ProxyGet(p.xl, host, p.eblock, p.from, p.to)
		}
		if err == nil {
			r.Close()
			p.mu.Lock()
			closed := p.closed
			// 如果在p.stg.Get()过程中, p被调用了Close()了, 需要自己关掉自己的句柄
			if closed {
				rc.Close()
			}
			p.rc = rc
			p.mu.Unlock()
			if n == 0 {
				return p.Read(b)
			}
			break
		}
		if p.proxy {
			if e, ok := err.(rpc.RespError); (ok && e.HttpCode() == 503) || (!ok && strings.Contains(err.Error(), "connecting to proxy")) {
				p.xl.Error("all proxy fails", err)
				err = errors.Info(err, "stg.Get", host).Detail(err)
				break
			}
		}
		p.trys++
		if p.trys >= nhost {
			err = errors.Info(err, "stg.Get", host).Detail(err)
			break
		}
		p.xl.Errorf("blockReader.Read: stg.Get %v failed => %v", host, err)
	}
	return n, err
}

func (p *blockReader) Close() error {
	p.mu.Lock()
	rc := p.rc
	p.closed = true
	p.mu.Unlock()
	return rc.Close()
}
