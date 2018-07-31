package ioutil

import (
	"bytes"
	"io"
)

// ----------------------------------------------------------------------------
// func ReadAll

// 为什么不直接用标准库 io/ioutil.ReadAll？
// 答：因为标准库中优先用 ReadFrom，而不是 WriteTo，而 fop.Source 对 WriteTo 作了优化。
func ReadAll(r io.Reader) (b []byte, err error) {

	buf := bytes.NewBuffer(nil)
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()

	// 为什么有此判断？
	// 答：因为 fop.Source 如果支持 WriteTo 则应该尽可能用 WriteTo，而不是 Read
	if wt, ok := r.(io.WriterTo); ok {
		_, err = wt.WriteTo(buf)
	} else {
		_, err = buf.ReadFrom(r)
	}
	return buf.Bytes(), err
}

// ----------------------------------------------------------------------------
