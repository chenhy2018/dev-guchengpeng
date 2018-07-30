package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	. "github.com/qiniu/ctype"
	"github.com/qiniu/log.v1"
	"qbox.us/bufio"
	"qbox.us/cc"
	cioutil "qbox.us/cc/ioutil"
	"qbox.us/errors"
	"qbox.us/largefile"
)

func main() {
	dataPath := flag.String("datapath", "./", "mq data path, default is current directory")
	chunkBits := flag.Int("chunkbits", 22, "data storage chunkbits, default is 22 (4M)")
	debugLevel := flag.Int("debuglevel", 2, "debug level, default is 2 (Warning)")
	flag.Parse()
	log.SetOutputLevel(*debugLevel)

	type offInfo struct {
		offTimeout int64
		offGet     int64
		offPut     int64
	}
	mqs := make(map[string]offInfo)

	fis, err := cioutil.ReadDir(*dataPath)
	if err != nil {
		log.Fatal("read mq datapath failed", err)
	}

	for _, fi := range fis {
		name := fi.Name()
		if checkValidMQ(name) == nil {
			path := *dataPath + name
			offTimeout, offGet, offPut, err2 := MQLen(path, uint(*chunkBits))
			if err2 != nil {
				log.Warnf("get mq len from %s failed - offGet:%d offPut:%d err:%v", path, offGet, offPut, errors.Detail(err2))
				continue
			}
			mqs[name] = offInfo{offTimeout: offTimeout, offGet: offGet, offPut: offPut}
		}
	}

	// report result
	for name, info := range mqs {
		fmt.Printf("%-25s%d\t%d\t%d\t%d\n", name, info.offPut-info.offGet, info.offTimeout, info.offGet, info.offPut)
	}
}

func checkValidMQ(name string) (err error) {
	pos := strings.Index(name, "-")
	if pos <= 0 {
		err = errors.Info(errors.EINVAL, "invalid mq name")
		return
	}

	if !IsType(DIGIT|ALPHA, name[:pos]) {
		err = errors.Info(errors.EINVAL, "invalid mq name - invalid uid part")
		return
	}

	if !IsType(XMLSYMBOL_NEXT_CHAR, name[pos+1:]) {
		err = errors.Info(errors.EINVAL, "invalid mq name - invalid mqid part")
	}
	return
}

const headerSize = 8

const (
	stateNormal     = 0x01
	stateProcessing = 0x02
	stateDone       = 0x04

	stateTravelPut     = stateNormal | stateProcessing | stateDone
	stateTravelGet     = stateProcessing | stateDone
	stateTravelTimeout = stateDone
)

type indexFile struct {
	OffPut     int64
	OffGet     int64
	OffTimeout int64
	Expires    uint32
	Reserved   uint32
}

type recordHeader struct {
	MsgLen   uint16
	Tag      uint8
	State    uint8
	LockTime uint32 // in s
}

func MQLen(name string, chunkBits uint) (offTimeout, offGet, offPut int64, err error) {
	mqf, err := largefile.Open(name, chunkBits)
	if err != nil {
		err = errors.Info(err, "open data path failed -", name).Detail(err)
		return
	}

	findex, err := os.Open(name + "/INDEX")
	if err != nil {
		err = errors.Info(err, "open index file failed -", name+"/index").Detail(err)
		return
	}
	defer findex.Close()

	var index indexFile
	err = binary.Read(findex, binary.LittleEndian, &index)
	if err != nil {
		if err != io.EOF {
			err = errors.Info(err, "read index file failed -", name+"/index").Detail(err)
			return
		}
	}

	index.OffPut, err = travel(mqf, index.OffPut, stateTravelPut)
	if err != nil {
		errors.Info(err, "mq.NewInstance: travelPut failed").Detail(err).Warn()
		//TODO: truncate mqf to index.OffPut
	}

	index.OffTimeout, err = travel(mqf, index.OffTimeout, stateTravelTimeout)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: travelTimeout failed").Detail(err)
		return
	}

	index.OffGet, err = travel(mqf, index.OffGet, stateTravelGet)
	if err != nil {
		err = errors.Info(err, "mq.NewInstance: travelGet failed").Detail(err)
		return
	}
	log.Info("index:", index.OffPut, index.OffGet, index.OffTimeout)

	return index.OffTimeout, index.OffGet, index.OffPut, nil
}

func travel(mqf *largefile.Instance, off int64, allowStates uint8) (offNew int64, err error) {

	log.Debug("travel:", off, int(allowStates))

	const BufSize = 1 << 16

	f := &cc.Reader{ReaderAt: mqf, Offset: off}
	bufr := bufio.NewReaderSize(f, BufSize)

	for {
		var h recordHeader
		err = binary.Read(bufr, binary.LittleEndian, &h)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if (h.State & allowStates) == 0 {
			break
		}
		_, err = bufr.Next(int(h.MsgLen))
		if err != nil {
			break
		}
		off += (headerSize + int64(h.MsgLen))
	}
	return off, err
}
