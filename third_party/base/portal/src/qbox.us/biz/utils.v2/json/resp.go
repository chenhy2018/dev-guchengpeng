package json

import (
	"fmt"
)

const (
	codeOK = 200
)

type CommonResponse struct {
	Code int `json:"code"`
}

func (c *CommonResponse) Error() error {
	if c.Code != codeOK {
		return fmt.Errorf("request failed. code: %d", c.Code)
	}

	return nil
}
