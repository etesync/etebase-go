package etebase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientSignup(t *testing.T) {
	c := NewClient("https://api.etebase.com/developer/gchaincl/")
	assert.NoError(t,
		c.Signup("gchaincl", "foo"),
	)
}
