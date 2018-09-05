package mgo3

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify.v1/require"
)

func TestGetMongoHosts(t *testing.T) {
	cases := map[string][]string{
		"192.168.1.1":             []string{"192.168.1.1"},
		"192.168.1.1:27017":       []string{"192.168.1.1:27017"},
		"localhost":               []string{"localhost"},
		"192.168.1.1,192.168.1.2": []string{"192.168.1.1", "192.168.1.2"},

		"mongodb://192.168.1.1":             []string{"192.168.1.1"},
		"mongodb://192.168.1.1:27017":       []string{"192.168.1.1:27017"},
		"mongodb://localhost":               []string{"localhost"},
		"mongodb://192.168.1.1,192.168.1.2": []string{"192.168.1.1", "192.168.1.2"},

		"user:pass@192.168.1.1":             []string{"192.168.1.1"},
		"user:pass@192.168.1.1:27017":       []string{"192.168.1.1:27017"},
		"user:pass@localhost":               []string{"localhost"},
		"user:pass@192.168.1.1,192.168.1.2": []string{"192.168.1.1", "192.168.1.2"},

		"mongodb://user:pass@192.168.1.1":             []string{"192.168.1.1"},
		"mongodb://user:pass@192.168.1.1:27017":       []string{"192.168.1.1:27017"},
		"mongodb://user:pass@localhost":               []string{"localhost"},
		"mongodb://user:pass@192.168.1.1,192.168.1.2": []string{"192.168.1.1", "192.168.1.2"},

		"mongodb://user:pass@192.168.1.1/db":             []string{"192.168.1.1"},
		"mongodb://user:pass@192.168.1.1:27017/db":       []string{"192.168.1.1:27017"},
		"mongodb://user:pass@localhost/db":               []string{"localhost"},
		"mongodb://user:pass@192.168.1.1,192.168.1.2/db": []string{"192.168.1.1", "192.168.1.2"},

		"mongodb://user:pass@192.168.1.1/db?connect=direct":             []string{"192.168.1.1"},
		"mongodb://user:pass@192.168.1.1:27017/db?connect=direct":       []string{"192.168.1.1:27017"},
		"mongodb://user:pass@localhost/db?connect=direct":               []string{"localhost"},
		"mongodb://user:pass@192.168.1.1,192.168.1.2/db?connect=direct": []string{"192.168.1.1", "192.168.1.2"},

		"192.168.1.1/db?connect=direct":             []string{"192.168.1.1"},
		"192.168.1.1:27017/db?connect=direct":       []string{"192.168.1.1:27017"},
		"localhost/db?connect=direct":               []string{"localhost"},
		"192.168.1.1,192.168.1.2/db?connect=direct": []string{"192.168.1.1", "192.168.1.2"},

		"192.168.1.1?connect=direct":             []string{"192.168.1.1"},
		"192.168.1.1:27017?connect=direct":       []string{"192.168.1.1:27017"},
		"localhost?connect=direct":               []string{"localhost"},
		"192.168.1.1,192.168.1.2?connect=direct": []string{"192.168.1.1", "192.168.1.2"},
	}
	for raw, expect := range cases {
		log.Println(raw, expect)
		hosts, user, password, authDB := getMongoHosts(raw)
		require.Equal(t, expect, hosts)
		if user != "" {
			require.Equal(t, "user", user)
			require.Equal(t, "pass", password)
		}
		if strings.Contains(raw, "/db") {
			require.Equal(t, "db", authDB)
		}
	}
}
