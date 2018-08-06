package oauth

type Code int

const (
	CodeBadToken     Code = 40001
	CodeNoPermisson  Code = 40002
	CodeInvalidScope Code = 40003
)

var codesMap = map[Code]string{
	CodeBadToken:     "bad token",
	CodeNoPermisson:  "permission denied",
	CodeInvalidScope: "invalid scopes",
}

func (c Code) String() string {
	return codesMap[c]
}
