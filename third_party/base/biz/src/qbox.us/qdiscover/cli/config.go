package cli

import (
	"os"

	"qbox.us/qdiscover/discover"
)

const DiscoverHostEnvVar = "DISCOVERD_HOST"

var (
	DiscoverHost = "http://localhost:18888"
	Client       *discover.Client
)

func init() {
	if host := os.Getenv("DISCOVERD_HOST"); host != "" {
		DiscoverHost = host
	}
	Client = discover.New([]string{DiscoverHost}, nil)
}
