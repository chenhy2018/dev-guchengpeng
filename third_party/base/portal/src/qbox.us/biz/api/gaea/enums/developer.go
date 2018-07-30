package enums

type LicenseVersion string

const (
	LicenseNoFound LicenseVersion = ""
	LicenseActived LicenseVersion = "0.1"
)

type Gender int

const (
	GenderMale   Gender = 0
	GenderFemale Gender = 1
)

func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "先生"
	case GenderFemale:
		return "女士"
	default:
		return ""
	}
}
