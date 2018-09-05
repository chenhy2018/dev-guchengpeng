package urlquery

import (
	"net/url"
	"strings"

	"github.com/qiniu/ctype"
)

func ParseQuery(query string) (m url.Values, cmd string, err error) {
	m = make(url.Values)
	cmd, err = parseQuery(m, query)
	return
}

func parseQuery(m url.Values, query string) (cmd string, err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		i := strings.Index(key, "=")
		if i == -1 {
			cmd, err = url.QueryUnescape(key)
			if err != nil {
				return
			}
			continue
		}
		if !ctype.IsType(ctype.URLQUERY, key[:i]) {
			cmd, err = url.QueryUnescape(key)
			if err != nil {
				return
			}
			continue
		}
		key, value = key[:i], key[i+1:]
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			err = err1
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			err = err1
			continue
		}
		m[key] = append(m[key], value)
	}
	return
}
