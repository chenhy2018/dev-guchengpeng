package ios

import (
	"encoding/base64"
	"fmt"
	"github.com/qiniu/log.v1"
	"strings"
	"testing"
	//	"github.com/qiniu/ts"
)

func init() {
	log.SetOutputLevel(0)
}

func doTestSha1Calcer(data string, max int, chunkSize int, times int, t *testing.T) {

	fmt.Println("-----------------------")
	runner := Goroutine{}
	f := strings.NewReader(data)
	c := NewSha1Calcer(f, runner, max, chunkSize)
	for {
		if times == 0 {
			c.Cancel(nil)
			break
		}
		times--
		keys, err := c.Get()
		if len(keys) != 0 {
			fmt.Println(len(keys)/20, base64.URLEncoding.EncodeToString(keys))
		}
		if err != nil {
			break
		}
	}
}

func TestSha1Calcer(t *testing.T) {

	doTestSha1Calcer("012345678901234567", 1, 4, -1, t)
	doTestSha1Calcer("012345678901234567", 2, 4, -1, t)
	doTestSha1Calcer("012345678901234567", 3, 4, -1, t)

	doTestSha1Calcer("012345678901234567", 1, 4, 1, t)
	doTestSha1Calcer("012345678901234567", 2, 4, 1, t)
	doTestSha1Calcer("012345678901234567", 3, 4, 0, t)
}
