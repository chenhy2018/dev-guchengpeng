package resource

import (
	"errors"
	"strings"
)

const (
	QRNPrefix    = "qrn"
	QRNSeparator = ":"
	QRNWildcard  = "*"
	QRNPartCount = 5
)

type QRN struct {
	Product  string
	Zone     string
	UID      string
	Resource string
}

var ErrBadQRN = errors.New("bad qrn")

func ParseQRN(str string) (*QRN, error) {
	var subs []string
	switch {
	case strings.HasPrefix(str, QRNPrefix+QRNSeparator):
		subs = strings.SplitN(str, QRNSeparator, QRNPartCount)
		if len(subs) != QRNPartCount ||
			subs[0] != QRNPrefix ||
			subs[1] == "" || subs[4] == "" {
			return nil, ErrBadQRN
		}
	case str == "*":
		subs = []string{QRNPrefix, "*", "*", "*", "*"}
	default:
		return nil, ErrBadQRN
	}
	return &QRN{
		Product:  strings.TrimSpace(subs[1]),
		Zone:     strings.TrimSpace(subs[2]),
		UID:      strings.TrimSpace(subs[3]),
		Resource: strings.TrimSpace(subs[4]),
	}, nil
}

func (q *QRN) Parts() []string {
	return []string{QRNPrefix, q.Product, q.Zone, q.UID, q.Resource}
}

func (q *QRN) String() string {
	return strings.Join(q.Parts(), QRNSeparator)
}
