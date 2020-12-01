package etebase

import (
	"fmt"
)

type User struct {
	Username string `msgpack:"username"`
	Email    string `msgpack:"email"`
}

type ErrorResponse struct {
	Code   string          `msgpack:"code"`
	Detail string          `msgpack:"detail"`
	Errors []ErrorResponse `msgpack:"errors,omitempty"`
}

func (err *ErrorResponse) Error() string {
	return fmt.Sprintf("code: %s, detail: %s", err.Code, err.Detail)
}

type SignupRequest struct {
	User             User   `msgpack:"user"`
	Salt             []byte `msgpack:"salt"`
	LoginPubkey      []byte `msgpack:"loginPubkey"`
	PubKey           []byte `msgpack:"pubkey"`
	EncryptedContent []byte `msgpack:"encryptedContent"`
}

type LoginChallengeRequest struct {
	Username string `msgpack:"username"`
}

type LoginChallengeResponse struct {
	Salt      []byte `msgpack:"salt"`
	Challenge []byte `msgpack:"challenge"`
}

type LoginRequest struct {
	// These fields are common to login and passwordChange
	Username  string `msgpack:"username"`
	Challenge []byte `msgpack:"challenge"`
	Host      string `msgpack:"host"`
	Action    string `msgpack:"action"`

	// These fields exclusively used for passwordChange
	LoginPubKey      []byte `msgpack:"loginPubkey,omitempty"`
	EncryptedContent []byte `msgpack:"encryptedContent,omitempty"`
}

type LoginResponse struct {
	Token string `msgpack:"token"`
	User  struct {
		User
		PubKey           string `msgpack:"pubkey"`
		EncryptedContent string `msgpack:"encryptedContent"`
	} `msgpack:"user"`
}
