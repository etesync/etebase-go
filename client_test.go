package etebase

import (
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
		require.NoError(t,
			acc.Login(user.Username, password),
		)
	})

	t.Run("Play", func(t *testing.T) {
		require.NoError(t, acc.Play())
	})
}
