package templatefunc

import (
	"html/template"
)

func Raw(text string) template.HTML {
	return template.HTML(text)
}
