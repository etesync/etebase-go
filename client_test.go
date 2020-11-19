package etebase

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	var (
		acc = NewAccount(
			NewClient("gchaincl", DefaultClientOptions),
		)

		user = User{
			Username: "gchaincl",
			Email:    "gchain@pm.me",
		}
		password = "foo"
	)

	t.Run("Signup", func(t *testing.T) {
		assert.NoError(t,
			acc.Signup(user, password),
		)
	})

	t.Run("Login", func(t *testing.T) {
		resp, err := acc.Login(user.Username, password)
		require.NoError(t, err)
		log.Printf("resp = %+v\n", resp.Token)
	})
}
