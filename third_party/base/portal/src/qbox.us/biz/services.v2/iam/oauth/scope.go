package oauth

type Scope string

const (
	ScopeUserProfile  Scope = "user_profile"
	ScopeUserKeypairs Scope = "user_keypairs"
)

func (s Scope) String() string {
	return string(s)
}
