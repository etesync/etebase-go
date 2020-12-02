package etebase

import (
	"log"
	"net/http"

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

// NewAccount returns a new Account object.
func NewAccount(c *Client) *Account {
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

// IsEtebaseServer checks if the provided client is pointing to an actual
// Etebase server, it returns false if not.
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

	if resp.StatusCode != http.StatusCreated {
		var respErr ErrorResponse
		if err := codec.NewDecoder(resp.Body).Decode(&respErr); err != nil {
			return err
		}
		return &respErr
	}

	var loginResponse LoginResponse
	if err := codec.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	acc.session = &loginResponse

	return nil
}

// Login attempts to login a user given its username and password.
// If login succeeds, a session will be created.
// The account will be authentified and ready to perform requests
// on behalf of that user.
func (acc *Account) login(username, password string) error {
	lc, err := acc.loginChallenge(username)
	if err != nil {
		return err
	}
	acc.initKeys(lc.Salt, password)

	req := LoginRequest{
		Username:  username,
		Challenge: lc.Challenge,
		Host:      "api.etebase.com",
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

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp ErrorResponse
		if err := codec.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return err
		}
		return &errResp
	}

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
	dec := codec.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var rErr ErrorResponse
		if err := dec.Decode(&rErr); err != nil {
			return nil, err
		}
		return nil, &rErr
	}

	var challenge LoginChallengeResponse
	if err := dec.Decode(&challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

// PasswordChange changes the password of an active session.
func (acc *Account) PasswordChange(newPassword string) error {
	if acc.session == nil {
		return ErrNoSession
	}
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
		Host:             "api.etebase.com",
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
	defer resp.Body.Close()

	var loginResponse interface{} //LoginResponse
	if err := codec.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	return nil
}

// Collection is not implemented yet.
func (acc *Account) Collection() error {
	if acc.session == nil {
		return ErrNoSession
	}

	resp, err := acc.client.WithToken(acc.session.Token).Post("/collection/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("resp.Status = %+v\n", resp.Status)

	var body interface{}
	if err := codec.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	log.Printf("body = %+v\n", body)

	return err
}

func Signup(c *Client, user User, password string) (*Account, error) {
	acc := NewAccount(c)
	if err := acc.signup(user, password); err != nil {
		return nil, err
	}
	return acc, nil
}

func Login(c *Client, username, password string) (*Account, error) {
	acc := NewAccount(c)
	if err := acc.login(username, password); err != nil {
		return nil, err
	}
	return acc, nil
}
