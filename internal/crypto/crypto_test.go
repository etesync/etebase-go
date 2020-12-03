// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoRand(t *testing.T) {
	r1 := Rand(32)
	r2 := Rand(32)
	require.NotEqual(t, r1, r2)
}

func TestCryptoDeriveKey(t *testing.T) {
	var (
		rand = Rand(32)
		pass = "password"
	)

	key1 := DeriveKey(rand, pass)
	key2 := DeriveKey(rand, pass)
	require.Equal(t, key1, key2)

	key3 := DeriveKey(Rand(32), pass)
	require.NotEqual(t, key1, key3)

}

func TestCrytoEncrypt(t *testing.T) {
	var (
		key = []byte("secret key")
		msg = []byte("the message")
	)

	enc, err := Encrypt(key, msg)
	require.NoError(t, err)

	got, err := Decrypt(key, enc)
	require.NoError(t, err)

	assert.EqualValues(t, msg, got)
}
