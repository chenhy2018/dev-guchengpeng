package ios

import (
	"fmt"
	"io"
	"log"
	"qbox.us/rpc"
	"strconv"
	//	"qbox.us/cc"
	"qbox.us/api"
)

// ----------------------------------------------------------

func (c Channel) UploadEx(
	data interface{}, callback string, f io.Reader, fsize int64, mode int) (code int, err error) {

	windowSize := c.Io.WindowSize

	if fsize <= windowSize {
		_, code, err = c.CreateFileEx(data, callback, fsize, f, mode)
	} else {
		var ret CreateFileRet
		ret, code, err = c.CreateFileEx(nil, callback, fsize, io.LimitReader(f, windowSize), mode)
		for off := windowSize; off < fsize; off += windowSize {
			code, err = c.WriteAt(data, ret.Fid, off, io.LimitReader(f, windowSize))
			if err != nil {
				return
			}
		}
	}
	return
}

// ----------------------------------------------------------

func (c Channel) CreateFid() string {
	c.fidBase++
	return strconv.Itoa(c.fidBase)
}

func (c Channel) OpenFileEx(
	fid string, data interface{}, callback string,
	fsize int64, first io.Reader, mode int) (code int, err error) {

	ret := writeAtRet{Data: data}
	url := c.Io.Host + "/chan/" + c.Id + "/create/" + rpc.EncodeURI(callback) +
		"/fid/" + fid + "/fsize/" + strconv.FormatInt(fsize, 10)
	if mode != 0 {
		url += "/mode/" + strconv.Itoa(mode)
	}
	code, err = c.Io.Conn.CallWithBinary1(&ret, url, first)
	return
}

func (c Channel) OpenFileWithKeys(
	fid string, data interface{}, callback string,
	fsize int64, first []byte, mode int) (required int64, code int, err error) {

	ret := writeAtRet{Data: data}
	url := c.Io.Host + "/chan/" + c.Id + "/createWithKeys/" + rpc.EncodeURI(callback) +
		"/fid/" + fid + "/fsize/" + strconv.FormatInt(fsize, 10)
	if mode != 0 {
		url += "/mode/" + strconv.Itoa(mode)
	}
	code, err = c.Io.Conn.CallWithParam(&ret, url, "application/octet-stream", first)
	required = ret.Required
	return
}

type UploadParams struct {
	Runner      Runner
	Sha1KeysMax int
	ChunkSize   int
	Callback    string
	Fsize       int64
	Mode        int
	Retrying    bool
	PutWithKeys bool
}

func (r *UploadParams) InitFastUpload(runner Runner, maxKeys int, chunkSize int) {

	r.Runner = runner
	r.Sha1KeysMax = maxKeys
	r.ChunkSize = chunkSize
	r.PutWithKeys = true
}

func (c Channel) tryFastUpload(
	fid string, data interface{}, f io.ReadSeeker, params *UploadParams) (code int, err error) {

	if params.ChunkSize == 0 {
		log.Println("FastUpload failed: invalid chunkSize")
		return api.InvalidArgs, api.EInvalidArgs
	}

	code = api.FunctionFail
	calcer := NewSha1Calcer(f, params.Runner, params.Sha1KeysMax, params.ChunkSize)
	keys, err := calcer.Get()
	if len(keys) == 0 {
		log.Println("FastUpload failed:", err)
		return
	}

	required, code2, err2 := c.OpenFileWithKeys(
		fid, data, params.Callback, params.Fsize, keys, params.Mode)
	if err2 != nil {
		log.Println("FastUpload failed: OpenFileWithKeys -", err2)
		code, err = code2, err2
		return
	}

	if err == nil {
		chunkSize := int64(params.ChunkSize)
		excepted := int64(len(keys)/20) * chunkSize
		for required == excepted {
			keys, err = calcer.Get()
			if len(keys) > 0 {
				required, code2, err2 = c.WriteAtWithKeys(data, fid, required, keys)
				if err2 != nil {
					log.Println("FastUpload failed: WriteAtWithKeys -", err2)
					code, err = code2, err2
					return
				}
			}
			if err != nil {
				break
			}
			excepted += int64(len(keys)/20) * chunkSize
		}
	}
	if required == params.Fsize {
		return 200, nil
	}
	//	log.Println("required:", required)

	calcer.Cancel(err)
	_, err = f.Seek(required, 0)
	if err != nil {
		log.Println("FastUpload.Seek failed:", err)
		return
	}

	log.Println("SeekTo:", required)
	code, err = c.WriteAt(data, fid, required, f)
	if err == nil {
		return
	}

	log.Println("Channel.WriteAt failed:", err)
	return
}

var FastUploadMinFsize int64 = 16 * 1024

func (c Channel) ResumableUploadEx(
	fid string, data interface{}, f io.ReadSeeker, params *UploadParams) (code int, err error) {

	var ci ChannelInfo
	var serr error
	var retryTimes int

	if params.Fsize == 0 {
		return c.OpenFileEx(fid, data, params.Callback, 0, f, params.Mode)
	}

retry:
	if retryTimes > c.RetryTimes {
		return
	}
	retryTimes++

	if !params.Retrying {
		if params.PutWithKeys && params.Fsize >= FastUploadMinFsize {
			code, err = c.tryFastUpload(fid, data, f, params)
			if err == nil {
				return
			}
		} else {
			code, err = c.OpenFileEx(fid, data, params.Callback, params.Fsize, f, params.Mode)
			if err == nil {
				return
			}
			log.Println("Channel.OpenFileEx failed:", err)
			params.PutWithKeys = true
		}
		if code == InvalidChan {
			goto refresh
		}
		if err == EIoCanceled || err == EIoUnresumable {
			return
		}
	}

	ci, code, serr = c.Stat()
	if serr != nil {
		log.Println("Channel.Stat failed:", err)
		if code == InvalidChan {
			goto refresh
		}
		if code != NoFile {
			err = serr
		}
		return
	}
	if ci.Fid != fid || ci.Callback != params.Callback || ci.Fsize != params.Fsize {
		log.Println("Channel.Stat failed: data not match")
		params.Retrying = false
		goto retry
	}

	_, err = f.Seek(ci.Pos, 0)
	if err != nil {
		log.Println("ResumableUploader.Seek failed:", err)
		code = api.FunctionFail
		return
	}

	log.Println("ResumableUploader.SeekTo:", ci.Pos)
	code, err = c.WriteAt(data, fid, ci.Pos, f)
	if err == nil {
		return
	}

	log.Println("Channel.WriteAt failed:", err)
	if code == InvalidChan {
		goto refresh
	}
	if err == EIoCanceled || err == EIoUnresumable {
		return
	}
	params.Retrying = true
	goto retry

refresh:
	c.Id, code, err = c.Io.mkchan()
	if err != nil {
		log.Println("Channel.refresh failed:", err)
		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Println("ResumableUploader.Seek failed:", err)
		code = api.FunctionFail
		return
	}
	log.Println("ResumableUploader.SeekToBegin")
	params.Retrying = false
	goto retry
}

// ----------------------------------------------------------

func (c Channel) MetricResumableUploadEx(
	fid string, data interface{}, f io.ReadSeeker, params *UploadParams) (code int, err error) {

	mr := NewMetricReader(f, params.Fsize)
	code, err = c.ResumableUploadEx(fid, data, mr, params)
	if err == nil {
		fmt.Printf("\n%v KB/s\n", mr.AvgSpeed()/1000)
	}
	return
}

// ----------------------------------------------------------
