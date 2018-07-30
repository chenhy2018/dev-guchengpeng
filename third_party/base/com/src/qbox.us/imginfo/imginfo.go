package imginfo

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"io"
	"strconv"

	_ "code.google.com/p/go.image/tiff"
	_ "code.google.com/p/vp8-go/webp"
	"github.com/qiniu/log.v1"
	_ "golang.org/x/image/bmp"

	"qbox.us/errors"
	_ "qbox.us/jpeg"
)

type ImageConfig struct {
	Format     string `json:"format" bson:"format"`
	Width      int    `json:"width" bson:"width"`
	Height     int    `json:"height" bson:"height"`
	ColorModel string `json:"colorModel" bson:"colorModel"`
	FrameNum   int    `json:"frameNumber,omitempty" bson:"frameNumber,omitempty"`
}

func Get(r io.Reader) (ret ImageConfig, err error) {

	buf := &bytes.Buffer{}
	source := io.TeeReader(r, buf)
	cfg, format, err := image.DecodeConfig(source)
	if err != nil {
		ret.Format = format // format有可能是正确的，需要返回
		err = errors.Info(err, "image.DecodeConfig failed").Detail(err).Warn()
		return
	}

	var colorModel string
	switch cm := cfg.ColorModel.(type) {
	case color.Palette:
		colorModel = "palette" + strconv.Itoa(len(cm))
	default:
		switch cfg.ColorModel {
		case color.RGBAModel:
			colorModel = "rgba"
		case color.RGBA64Model:
			colorModel = "rgba64"
		case color.NRGBAModel:
			colorModel = "nrgba"
		case color.NRGBA64Model:
			colorModel = "nrgba64"
		case color.AlphaModel:
			colorModel = "alpha"
		case color.Alpha16Model:
			colorModel = "alpha16"
		case color.GrayModel:
			colorModel = "gray"
		case color.Gray16Model:
			colorModel = "gray16"
		case color.YCbCrModel:
			colorModel = "ycbcr"
		default:
			log.Warn("ColorModel:", cm)
		}
	}

	var imgConf *ImageConfig
	if format != "gif" {
		imgConf = &ImageConfig{format, cfg.Width, cfg.Height, colorModel, 0}
	} else {
		// gif的header没有包含frame number的信息，因此需要把所有image解码出来
		// 参见http://www.w3.org/Graphics/GIF/spec-gif89a.txt
		imgSource := io.MultiReader(buf, r)
		gifInfo, err1 := GetGifRichInfo(imgSource)
		if err1 != nil {
			err = errors.Info(err1, "gif.DecodeAll failed").Detail(err1)
			return
		}

		imgConf = &ImageConfig{format, cfg.Width, cfg.Height, colorModel, gifInfo.FrameNumber}
	}

	return *imgConf, nil
}
