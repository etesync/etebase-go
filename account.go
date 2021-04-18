// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package etebase

import (
	"github.com/etesync/etebase-go/internal/codec"
	"github.com/etesync/etebase-go/internal/crypto"
)

// Account represents a user account and is the main object for all user
// interactions and data manipulation.
type Account struct {
	client  *Client
	session *LoginResponse

	salt                []byte
	mainKey, accountKey []byte
	authPub, authPriv   []byte
	idPub, idPriv       []byte
}

// newAccount returns a new Account object.
func newAccount(c *Client) *Account {
	acc := &Account{
		client: c,
	}

	return acc
}

func (acc *Account) initKeys(salt []byte, password string) {
	acc.salt = salt
	acc.mainKey = crypto.DeriveKey(acc.salt, password)
	acc.accountKey = crypto.Rand(32)
	acc.authPub, acc.authPriv = crypto.GenrateKeyPair(acc.mainKey)
	acc.idPub, acc.idPriv = crypto.GenrateKeyPair(crypto.Rand(32))
}

// IsEtebaseServer checks whether the Client is pointing to a valid Etebase
// server.
func (acc *Account) IsEtebaseServer() (bool, error) {
	resp, err := acc.client.Get("/authentication/is_etebase/")
	if err != nil {
		return false, err
	}
	if err := resp.Body.Close(); err != nil {
		return false, err
	}

	return resp.StatusCode != 404, nil
}

func (acc *Account) signup(user User, password string) error {
	acc.initKeys(crypto.Rand(32), password)
	encrypedContent, err := crypto.Encrypt(acc.mainKey, append(acc.accountKey, acc.idPriv...))
	if err != nil {
		return err
	}

	body := SignupRequest{
		User:             user,
		Salt:             acc.salt,
		LoginPubkey:      acc.authPub,
		PubKey:           acc.idPub,
		EncryptedContent: encrypedContent,
	}
	resp, err := acc.client.Post("/authentication/signup", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResponse LoginResponse
	if err := codec.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	acc.session = &loginResponse
	return nil
}

func (acc *Account) login(username, password string) error {
	lc, err := acc.loginChallenge(username)
	if err != nil {
		return err
	}
	acc.initKeys(lc.Salt, password)

	req := LoginRequest{
		Username:  username,
		Challenge: lc.Challenge,
		Host:      acc.client.Host(),
		Action:    "login",
	}
	buf, err := codec.Marshal(req)
	if err != nil {
		return err
	}

	sig := crypto.Sign(acc.authPriv, buf)
	resp, err := acc.client.Post("/authentication/login", struct {
		Response  []byte `msgpack:"response"`
		Signature []byte `msgpack:"signature"`
	}{buf, sig})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResponse LoginResponse
	if err := codec.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	acc.session = &loginResponse
	return nil
}

func (acc *Account) loginChallenge(username string) (*LoginChallengeResponse, error) {
	resp, err := acc.client.Post("/authentication/login_challenge", &LoginChallengeRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var challenge LoginChallengeResponse
	if err := codec.NewDecoder(resp.Body).Decode(&challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

// PasswordChange changes the password of an active session.
func (acc *Account) PasswordChange(newPassword string) error {
	lc, err := acc.loginChallenge(acc.session.User.Username)
	if err != nil {
		return err
	}

	priv := acc.authPriv
	acc.initKeys(lc.Salt, newPassword)

	encrypedContent, err := crypto.Encrypt(acc.mainKey, append(acc.accountKey, acc.idPriv...))
	if err != nil {
		return err
	}

	req := LoginRequest{
		Username:         acc.session.User.Username,
		Challenge:        lc.Challenge,
		Host:             acc.client.Host(),
		Action:           "changePassword",
		LoginPubKey:      acc.authPub,
		EncryptedContent: encrypedContent,
	}
	buf, err := codec.Marshal(req)
	if err != nil {
		return err
	}

	sig := crypto.Sign(priv, buf)
	resp, err := acc.client.WithToken(acc.session.Token).Post("/authentication/change_password", struct {
		Response  []byte `msgpack:"response"`
		Signature []byte `msgpack:"signature"`
	}{buf, sig})
	if err != nil {
		return err
	}

	// We don't expect any content from the server.
	return resp.Body.Close()
}

// Logout the user from the current session and invalidate the authentication
// token.
func (acc *Account) Logout() error {
	resp, err := acc.client.WithToken(acc.session.Token).Post("/authentication/logout/", nil)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// Collection is not implemented yet.
func (acc *Account) CreateCollection(col *EncryptedCollection) error {
	resp, err := acc.client.WithToken(acc.session.Token).Post("/collection", col)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body interface{}
	if err := codec.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	return err
}

func (acc *Account) GetCollection(id string) (*EncryptedCollection, error) {
	resp, err := acc.client.WithToken(acc.session.Token).Get("/collection/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body EncryptedCollection
	if err := codec.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}

func (acc *Account) ListCollections(types [][]byte) ([]EncryptedCollection, error) {
	req := struct {
		CollectionTypes [][]byte `msgpack:"collectionTypes"`
	}{
		types,
	}

	resp, err := acc.client.WithToken(acc.session.Token).Post("/collection/list_multi", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		Data []EncryptedCollection `msgpack:"data"`
	}
	if err := codec.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Data, nil
}

// Signup a new user account and returns an Account instance.
func Signup(c *Client, user User, password string) (*Account, error) {
	acc := newAccount(c)
	if err := acc.signup(user, password); err != nil {
		return nil, err
	}
	return acc, nil
}

// Login a user and returns a handle to an Account instance.
func Login(c *Client, username, password string) (*Account, error) {
	acc := newAccount(c)
	if err := acc.login(username, password); err != nil {
		return nil, err
	}
	return acc, nil
}
