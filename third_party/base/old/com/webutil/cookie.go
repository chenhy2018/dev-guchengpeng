package webutil

import (
	"http"
	"os"
)

func Cookie(req *http.Request) (m http.Values, err os.Error) {
	return ParseCookie(req.Header["Cookie"])
}

func ParseCookie(values []string) (m http.Values, err os.Error) {
	m = http.Values{}
	for _, s := range values {
		key := ""
		begin := 0
		end := 0
		for i := 0; i < len(s); i++ {
			switch s[i] {
			case ' ', '\t':
				// leading whitespace?
				if begin == end {
					begin = i + 1
					end = begin
				}
			case '=':
				if key == "" {
					key = s[begin:end]
					begin = i + 1
					end = begin
				} else {
					end += 1
				}
			case ';':
				if len(key) > 0 && begin < end {
					value := s[begin:end]
					m.Add(key, value)
				}
				key = ""
				begin = i + 1
				end = begin
			default:
				end = i + 1
			}
		}
		if len(key) > 0 && begin < end {
			m.Add(key, s[begin:end])
		}
	}
	return
}
