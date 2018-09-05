// imageAve，取图片平均色调,请求为: url?imageAve，返回结果为: {"RGBA":"0x7e7e7f"}
// http://pm.qbox.me/redmine/issues/6365
package imgave

import (
	"fmt"
	"image"
	"io"
)

// support image types
import (
	_ "golang.org/x/image/bmp"
	_ "code.google.com/p/vp8-go/webp"
	_ "image/gif"
	_ "image/png"
	_ "qbox.us/jpeg"
)

//归约为30w相素进行计算，实际计算的尺寸略大于此值
var mAXPIXELCOUNT = uint32(300000)

const (
	Gr = 13933
	Gg = 46971
	Gb = 4732
)

type ImageAve struct {
	R   uint8  `json:"-"`
	G   uint8  `json:"-"`
	B   uint8  `json:"-"`
	RGB string `json:"RGB" bson:"RGB"`
}

func (i *ImageAve) doHexString() {
	rgb := int(i.R)<<16 | int(i.G)<<8 | int(i.B)
	i.RGB = fmt.Sprintf("0x%06x", rgb)
}

func Get(reader io.Reader) (ave ImageAve, err error) {

	m, _, err := image.Decode(reader)
	if err != nil {
		return
	}
	ave = getImageAve(m)
	ave.doHexString()
	return
}

func getImageAve(m image.Image) (ave ImageAve) {

	bounds := m.Bounds()
	height, width := (bounds.Max.Y - bounds.Min.Y), (bounds.Max.X - bounds.Min.X)
	wh := uint32(width * height)
	var sr, sg, sb, count, w uint32
	step := 1
	if wh > mAXPIXELCOUNT {
		step = int(wh / mAXPIXELCOUNT)
	}

	for i := 0; i < height*width; i += step {
		x := i % width
		y := i / width
		r, g, b, a := m.At(x, y).RGBA()
		if a != 0 {
			r, g, b = r>>8, g>>8, b>>8
			//加权
			w = 1
			if iGray := (Gr*r + Gg*g + Gb*b) >> 16; iGray > 50 && iGray < 200 {
				w = 2
			}
			sr, sg, sb = sr+r*w, sg+g*w, sb+b*w
			count += w
		}
	}

	if count == 0 {
		return
	}

	ave.R, ave.G, ave.B = _tr(sr/count), _tr(sg/count), _tr(sb/count)
	return
}

func _tr(val uint32) uint8 {
	if val > 255 {
		return 255
	}
	return uint8(val)
}
