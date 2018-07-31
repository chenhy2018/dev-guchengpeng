package image_preview

import (
	"encoding/json"
	"strings"
)

// ------------------------------------------------------------------------------------------

type PropVal struct {
	Width, Height, Quality, Mode int
	Sharpen                      string
}

func Marshal(v *PropVal) (data string, err error) {

	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	data = string(b)
	return
}

func Unmarshal(data string, v *PropVal) (err error) {

	in := strings.NewReader(data)
	err = json.NewDecoder(in).Decode(v)
	return
}

// ------------------------------------------------------------------------------------------
