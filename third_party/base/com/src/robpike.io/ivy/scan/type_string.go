// generated by stringer -type Type; DO NOT EDIT

package scan

import "fmt"

const _Type_name = "EOFErrorNewlineAssignCharCharConstantDefGreaterOrEqualIdentifierLeftBrackLeftParenNumberOperatorRationalRawStringRightBrackRightParenSemicolonSpaceString"

var _Type_index = [...]uint8{3, 8, 15, 21, 25, 37, 40, 54, 64, 73, 82, 88, 96, 104, 113, 123, 133, 142, 147, 153}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)) {
		return fmt.Sprintf("Type(%d)", i)
	}
	hi := _Type_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _Type_index[i-1]
	}
	return _Type_name[lo:hi]
}