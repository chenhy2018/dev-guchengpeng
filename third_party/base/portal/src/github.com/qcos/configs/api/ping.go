package api

import (
	"fmt"

	"code.google.com/p/go.net/context"
)

func (p *Client) Ping() (err error) {
	resp, er := p.Client.Do(context.Background(), "GET", "/ping", "", nil, 0)
	if resp != nil {
		defer resp.Body.Close()
	}
	if er != nil || resp.StatusCode != 200 {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		err = fmt.Errorf("host cannot connect: %v, %d", er, code)
		return
	}
	return
}
