package mailgun

import (
	"reflect"
	"testing"
)

// -----------------------------------------------------------------------------

func TestEmails(t *testing.T) {

	emails := Emails("x@qiniu.com;foo@a.com,bar@b.com")
	if !reflect.DeepEqual(emails, []string{"x@qiniu.com", "foo@a.com", "bar@b.com"}) {
		t.Fatal("Emails failed")
	}
}

// -----------------------------------------------------------------------------

