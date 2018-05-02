package conf

type argument struct {
	IP       *string
	Port     *string
	Realm    *string
}

var (
	Args     argument
)
