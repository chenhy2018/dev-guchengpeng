package types

import (
	"qbox.us/biz/utils.v2"
)

type Byte int64

func (b Byte) Humanize(precision ...int) string {
	p := 2
	if len(precision) > 0 {
		p = precision[0]
	}

	if b >= 1024*1024*1024*1024 {
		return utils.ToStr(float64(b)/1024/1024/1024/1024, p) + " T"
	}
	if b >= 1024*1024*1024 {
		return utils.ToStr(float64(b)/1024/1024/1024, p) + " G"
	}
	if b >= 1024*1024 {
		return utils.ToStr(float64(b)/1024/1024, p) + " M"
	}
	if b >= 1024 {
		return utils.ToStr(float64(b)/1024, p) + " K"
	}
	return utils.ToStr(float64(b), p) + " "
}

func (b Byte) String(precision ...int) string {
	return b.Humanize(precision...) + "B"
}

func (b Byte) BitPerSecondString(precision ...int) string {
	return b.Humanize(precision...) + "b/s"
}

func (b Byte) KB() KB {
	return KB(float64(b) / 1024)
}

func (b Byte) MB() MB {
	return MB(float64(b) / 1024 / 1024)
}
func (b Byte) GB() GB {
	return GB(float64(b) / 1024 / 1024 / 1024)
}

type GB float64

func (a GB) Byte() Byte {
	return Byte(int64(a * 1024 * 1024 * 1024))
}

type MB float64

func (a MB) Byte() Byte {
	return Byte(int64(a * 1024 * 1024))
}

type KB float64

func (a KB) Byte() Byte {
	return Byte(int64(a * 1024))
}
