package line

func ParseMeasurement(buf []byte) (int, []byte) {
	mBuf := make([]byte, 0)
	i, r := 0, 0
	for ; i < len(buf); i++ {
		if buf[i] == '\\' && i < len(buf)-1 {
			if buf[i+1] == ' ' || buf[i+1] == ',' {
				mBuf = append(mBuf, buf[r:i]...)
				mBuf = append(mBuf, buf[i+1])
				r = i + 2
				i++
				continue
			}
		}
		if buf[i] == ' ' || buf[i] == ',' {
			mBuf = append(mBuf, buf[r:i]...)
			break
		}
	}
	return i, mBuf
}
