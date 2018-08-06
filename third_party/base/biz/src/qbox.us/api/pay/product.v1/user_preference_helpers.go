package product

import (
	"qbox.us/api/pay/pay"
)

func (p *UserPreference) IsItemEnabled(name pay.Item) bool {
	if item, ok := p.Items[name]; ok {
		return item.Enabled
	} else {
		return true
	}
}
