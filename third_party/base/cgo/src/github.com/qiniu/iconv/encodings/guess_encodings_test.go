package encodings_test

import (
	"encoding/hex"
	"github.com/qiniu/iconv/encodings"
	"github.com/qiniu/log.v1"
	"strings"
	"testing"
)

// ---------------------------------------------------

type testCase struct {
	text     string
	encoding int
}

var gbkText1 = `
	2f 2f 0d 0a 2f 2f 20 b1 be ce c4 bc fe ca c7 d7
	d4 b6 af c9 fa b3 c9 b5 c4 0d 0a 2f 2f 20 d7 d6
	b7 fb ca b9 d3 c3 c6 b5 b6 c8 b1 ed a3 a8 d6 b5
	d4 bd d0 a1 a3 ac b1 ed ca be ca b9 d3 c3 c6 b5
	b6 c8 d4 bd b8 df a3 a9 0d 0a 2f 2f 0d 0a
`

func hexText(text string) string {

	cutset := " \t\r\n"
	for i := 0; i < len(cutset); i++ {
		text = strings.Replace(text, cutset[i:i+1], "", -1)
	}
	b, _ := hex.DecodeString(text)
	return string(b)
}

func Test(t *testing.T) {

	cases := []testCase{
		{"在标准库 path.Match 的基础上，允许 pattern 结尾处为 **，表示前缀匹配", 0},
		{hexText(gbkText1), 1},
		{"Hello, world!", 0},
	}

	log.SetOutputLevel(0)

	names := []string{"UTF8", "GBK"}
	cds, err := encodings.Open(names)
	if err != nil {
		t.Fatal("encodings.Open", err)
	}
	defer cds.Close()

	outbuf := make([]byte, 1024)
	for _, c := range cases {
		text := []byte(c.text)
		encoding, delta := cds.Guess(text, outbuf)
		if encoding != c.encoding {
			t.Fatal("cds.Guess failed:", c.text, c.encoding)
		}
		log.Info("cds.Guess:", names[encoding], delta)
		log.Info(cds.Item(encoding).ConvString(c.text))
	}
}

// ---------------------------------------------------
