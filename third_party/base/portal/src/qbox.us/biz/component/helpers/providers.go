package helpers

import (
	"github.com/teapots/params"
	"github.com/teapots/teapot"
	"qbox.us/biz/component/providers/client"
)

func LoadClassicProviders(tea *teapot.Teapot) {
	// params parser
	tea.Provide(params.ParamsParser())

	// for base service use
	tea.Provide(client.Transport())
}
