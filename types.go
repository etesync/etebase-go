package etebase

import "fmt"

type Error struct {
	Code   string  `json:"code"`
	Detail string  `json:"detail"`
	Errors []Error `json:"errors,omitempty"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("code: %s, detail: %s", err.Code, err.Detail)
}

type User struct {
	Username string `msgpack:"username"`
	Email    string `msgpack:"email"`
}

type Signup struct {
	User             User   `msgpack:"user"`
	Salt             []byte `msgpack:"salt"`
	LoginPubkey      []byte `msgpack:"loginPubkey"`
	PubKey           []byte `msgpack:"pubkey"`
	EncryptedContent []byte `msgpack:"encryptedContent"`
}
