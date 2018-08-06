package imginfo

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	gifFile = "frame.gif"
)

func TestGifFrame(t *testing.T) {

	config, err := getFromFile(gifFile)
	assert.Nil(t, err)
	assert.Equal(t, config.FrameNum, 8, "frame number should be 8")
}

func getFromFile(fname string) (ret ImageConfig, err error) {

	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	return Get(f)
}
