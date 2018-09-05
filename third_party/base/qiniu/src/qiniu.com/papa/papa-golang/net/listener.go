package net

import (
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kavu/go_reuseport"
	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
)

// NOTE: 当 fullAddr 格式为：
//	<proto>:<address>;<proto>:<address> 时，
//	自动使用 HAListener 开启监听
//
// 此规则对 ReusableEnsureListen 和 EnsureListen 有效

//=============================================

// 使用 SO_REUSEPORT 特性监听地址（对 unix 套接字无效）
//
func ReusableEnsureListen(fullAddr string) (net.Listener, error) {

	parts := strings.Split(fullAddr, ";")
	if len(parts) == 2 {
		return ListenHA(
			strings.TrimSpace(parts[0]),
			strings.TrimSpace(parts[1]),
			ReusableEnsureListen)
	}
	netw, addr := ParseIPCAddr(fullAddr, "tcp")
	if IsLocalIPC(fullAddr) {
		return EnsureListenLocalIPC(netw, addr)
	}
	return reuseport.NewReusablePortListener(netw, addr)
}

func EnsureListen(fullAddr string) (net.Listener, error) {

	parts := strings.Split(fullAddr, ";")
	if len(parts) == 2 {
		return ListenHA(
			strings.TrimSpace(parts[0]),
			strings.TrimSpace(parts[1]),
			EnsureListen)
	}
	netw, addr := ParseIPCAddr(fullAddr, "tcp")
	if IsLocalIPC(fullAddr) {
		return EnsureListenLocalIPC(netw, addr)
	}
	return net.Listen(netw, addr)
}

// 尽最大努力监听在本地路径上：
// 	1. 如果路径存在，则向该路径发起连接，试探该路径是否被他人监听
//	2. 如果试探性连接成功，则说明该路径地址确实被占用，则放弃监听
//	3. 如果试探性连接成功，则说明该路径地址虽存在但无人使用，则删除之并重新监听
//
func EnsureListenLocalIPC(network, path string) (net.Listener, error) {

	// 试探该地址是否正在监听
	conn, err := net.Dial(network, path)
	if err != nil {
		os.Remove(path)
	} else {
		conn.Close()
	}
	return net.Listen(network, path)
}

//-------------------------------------------------
// NOTE:
// 	组合使用 HAListener 和 HADialer 可以达到热重启效果
// 	当前支持 unix domain socket 和 tcp ip:port

type HAListener struct {
	Master string
	Backup string

	lock     sync.RWMutex
	listener net.Listener

	listenFunc func(string) (net.Listener, error)
}

func ListenHA(master, backup string,
	listen func(string) (net.Listener, error)) (*HAListener, error) {

	l := &HAListener{
		Master:     master,
		Backup:     backup,
		listenFunc: listen,
	}
	if l.listenFunc == nil {
		l.listenFunc = EnsureListen
	}

	lis, err := listen(master)
	if err == nil {
		log.Info("ListenHA: listening on master:", master)
		l.listener = lis
		return l, nil
	}
	log.Warnf("ListenHA: listen master(%s) failed: %v", master, err)

	lis, err = listen(backup)
	if err == nil {
		log.Info("ListenHA: listening on backup:", backup)
		l.listener = lis
		go l.enthrone()
		return l, nil
	}
	log.Errorf("ListenHA: listen backup(%s) failed: %v", backup, err)

	return nil, errors.New("all addresses not avaiable")
}

func (p *HAListener) enthrone() {

	var err error
	var lis net.Listener

	for {
		<-time.After(5e9)
		lis, err = p.listenFunc(p.Master)
		if err == nil {
			break
		}
		log.Warn("backup still not avaiable:", err)
	}
	old := p.listener
	p.lock.Lock()
	p.listener = lis
	p.lock.Unlock()

	log.Warn("backup listener closed:", old.Close())
	log.Info("switched to master:", p.Master)
}

func (p *HAListener) Accept() (net.Conn, error) {

	p.lock.RLock()
	lis := p.listener
	p.lock.RUnlock()

	conn, err := lis.Accept()
	if err == nil {
		return conn, err
	}
	log.Warn("accept error:", err)

	// try again
	p.lock.RLock()
	lis1 := p.listener
	p.lock.RUnlock()

	if lis1 == lis {
		return conn, err
	}
	log.Warn("switch occurred, try accept agagin:", lis1.Addr())
	return lis1.Accept()
}

func (p *HAListener) Addr() net.Addr {

	p.lock.Lock()
	defer p.lock.Unlock()

	return p.listener.Addr()
}

func (p *HAListener) Close() error {

	p.lock.Lock()
	defer p.lock.Unlock()

	return p.listener.Close()
}

//=============================================

type addr struct {
	network string
	address string
}

type HADialer struct {
	net.Dialer

	lock    sync.RWMutex
	backups map[string]*addr
}

func NewHADialer(d net.Dialer) *HADialer {

	return &HADialer{
		Dialer:  d,
		backups: make(map[string]*addr),
	}
}

func (p *HADialer) Dial(network, address string) (net.Conn, error) {

	conn, err := p.Dialer.Dial(network, address)
	if err == nil {
		return conn, err
	}
	log.Warnf("dial %s:%s - %v", network, address, err)

	p.lock.RLock()
	backup, ok := p.backups[network+":"+address]
	p.lock.RUnlock()

	if !ok {
		return nil, err
	}

	log.Info("trying backup:", backup)
	conn, err = p.Dialer.Dial(backup.network, backup.address)
	if err != nil {
		log.Warnf("dial backup(%s) error: %v", backup, err)
	}
	return conn, err
}

func (p *HADialer) SetBackup(master, backup string) {

	network, address := ParseIPCAddr(backup, "tcp")
	p.lock.Lock()
	defer p.lock.Unlock()

	p.backups[master] = &addr{network, address}
}

func (p *HADialer) CleanBackup(master string) {

	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.backups, master)
}
