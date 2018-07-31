package validator

import (
	"regexp"
	"strings"
)

var (
	remail = regexp.MustCompile(`(?i)^[a-z0-9_]+(\.?[a-z0-9-_+])*?@([a-zA-Z0-9]([a-zA-Z0-9\-]*?[a-zA-Z0-9])?\.)+[a-zA-Z]{2,20}$`)
)

// NOTE: account cannot support + char in email address!
func IsEmail(mail string) bool {
	return remail.MatchString(mail) && !strings.Contains(mail, "+")
}
