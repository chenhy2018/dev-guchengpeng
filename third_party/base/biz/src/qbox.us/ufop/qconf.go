package ufop

func QconfUfopID(ufop string) string {
	return "ufop:" + ufop
}

func QconfUappID(uapp string) string {
	return "uapp:" + uapp
}

func QconfDomainID(domain string) string {
	return "uappdomain:" + domain
}
