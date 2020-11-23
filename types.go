package etebase

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type User struct {
	Username string `msgpack:"username"`
	Email    string `msgpack:"email"`
}

type ErrorResponse struct {
	Code   string          `json:"code"`
	Detail string          `json:"detail"`
	Errors []ErrorResponse `json:"errors,omitempty"`
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

type Base64URL []byte

func (b *Base64URL) UnmarshalJSON(data []byte) (err error) {
	// remove '"' from the json data
	str := strings.Trim(string(data), "\"")

	*b, err = base64.RawURLEncoding.DecodeString(str)
	return err
}

type LoginChallengeRequest struct {
	Username string `msgpack:"username"`
}

type LoginChallengeResponse struct {
	Salt      Base64URL `json:"salt"`
	Challenge Base64URL `json:"challenge"`
}

type LoginRequest struct {
	Username  string `msgpack:"username"`
	Challenge []byte `msgpack:"challenge"`
	Host      string `msgpack:"host"`
	Action    string `msgpack:"action"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		User
		PubKey           string `json:"pubkey"`
		EncryptedContent string `json:"encryptedContent"`
	} `json:"user"`
}
