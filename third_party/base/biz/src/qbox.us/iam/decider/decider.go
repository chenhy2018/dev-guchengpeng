package decider

import (
	"qbox.us/iam/decider/resource"
	"qbox.us/iam/entity"
)

type Policy interface {
	GetIUID() uint32
	GetVersion() string
	GetStatments() []entity.Statement
	IsEnabled() bool
}

type Decider interface {
	Verify(policy Policy, action string, qrn *resource.QRN) bool
}
