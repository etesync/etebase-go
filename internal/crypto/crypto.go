// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
)

func Rand(len int) ([]byte, error) {
	buf := make([]byte, len)
	_, err := rand.Read(buf)
	return buf, err
}

func DeriveKey(salt []byte, password string) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		1, 64*1024, 4, 32, // sensible values
	)
}

func GenrateKeyPair(seed []byte) ([]byte, []byte) {
	priv := ed25519.NewKeyFromSeed(seed)
	return priv.Public().(ed25519.PublicKey), priv
}

func Sign(priv []byte, msg []byte) []byte {
	return ed25519.Sign(priv, msg)
}

func resizeChachaKey(key []byte) []byte {
	s := chacha20poly1305.KeySize
	l := len(key)
	if l >= s {
		return key[:s]
	}

	newKey := make([]byte, s)
	copy(newKey, key)
	return newKey
}

func Encrypt(key []byte, msg []byte) ([]byte, error) {
	key = resizeChachaKey(key)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the message and append the ciphertext to the nonce.
	return aead.Seal(nonce, nonce, msg, nil), nil
}

func Decrypt(key []byte, msg []byte) ([]byte, error) {
	key = resizeChachaKey(key)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	// Split the nonce and the cipher message
	nonce, cipher := msg[:aead.NonceSize()], msg[aead.NonceSize():]

	// Decrypt the message and check it wasn't tampered with.
	return aead.Open(nil, nonce, cipher, nil)
}
