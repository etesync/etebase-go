// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoRand(t *testing.T) {
	r1, err := Rand(32)
	require.Nil(t, err)
	r2, err := Rand(32)
	require.Nil(t, err)
	require.NotEqual(t, r1, r2)

	// Ensure that Rand is not using unseeded math/rand
	// Values taken from go1.19.4
	require.NotEqual(t, r1, []byte{82, 253, 252, 7, 33, 130, 101, 79, 22, 63, 95, 15, 154, 98, 29, 114, 149, 102, 199, 77, 16, 3, 124, 77, 123, 187, 4, 7, 209, 226, 198, 73})
}

func TestCryptoDeriveKey(t *testing.T) {
	r1, err := Rand(32)
	require.Nil(t, err)
	var (
		rand = r1
		pass = "password"
	)

	key1 := DeriveKey(rand, pass)
	key2 := DeriveKey(rand, pass)
	require.Equal(t, key1, key2)

	r2, err := Rand(32)
	require.Nil(t, err)
	key3 := DeriveKey(r2, pass)
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
