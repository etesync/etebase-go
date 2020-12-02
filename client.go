package etebase

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/etesync/etebase-go/internal/crypto"
)

var (
	ErrNoSession = errors.New("account has not started a session, use Signup or Login first")
)

// ClientOptions allow you control specific options of the client.
// Most of the users should use DefaultClientOptions when constructing the
// client.
type ClientOptions struct {
	// Host is the Etebase server host.
	Host string
	// Prefix is a string used as a prefix for requests.
	// Possible values are `/partner/your-username` or
	// `/developer/your-username` if your are using etebase.com server.
	// For local server leave it blank.
	Prefix string

	// UseSSL specifies is ssl should be used or not.
	UseSSL bool
}

func (opts ClientOptions) baseUrl() string {
	var schema string
	if opts.UseSSL {
		schema = "https"
	} else {
		schema = "http"
	}

	return schema + "://" + path.Join(opts.Host, opts.Prefix, "api/v1")
}

// DefaultClientOptions will make your client point to the official Etebase
// server in 'developer' mode.
func DeveloperClientOptions(name string) ClientOptions {
	return ClientOptions{
		Host:   "api.etebase.com",
		UseSSL: true,
		Prefix: path.Join("developer", name),
	}
}

func PartnerClientOptions(name string) ClientOptions {
	return ClientOptions{
		Host:   "api.etebase.com",
		UseSSL: true,
		Prefix: path.Join("partner", name),
	}
}

// Client implements the network client to use to interact with the Etebase
// server.
type Client struct {
	baseUrl string
	token   string
}

// NewClient returns a new client object given a name (your etebase account name),
// and options inside the ClientOptions struct.
func NewClient(opts ClientOptions) *Client {
	return &Client{
		baseUrl: opts.baseUrl(),
	}
}

// WithToken returns a client that attaches a `Authorization: Token <token>` to
// any request.
func (c Client) WithToken(token string) *Client {
	c.token = token
	return &c
}

func (c *Client) url(path string) string {
	url := c.baseUrl + path
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

// Post posts an encoded value `v` to the server.
// `v` will be encoded using msgpack format.
func (c *Client) Post(path string, v interface{}) (*http.Response, error) {
	log.Printf("POST %s", path)
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url(path), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("Accept", "application/msgpack")
	if t := c.token; t != "" {
		req.Header.Set("Authorization", "Token "+t)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) Get(path string) (*http.Response, error) {
	log.Printf("GET %s", path)
	return http.Get(c.url(path))
}

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

// Signup attempts to signup a new user into the server given a user and a
// password.
func (acc *Account) Signup(user User, password string) error {
	if acc.salt == nil {
		acc.initKeys(crypto.Rand(32), password)
	}

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
		if err := NewDecoder(resp.Body).Decode(&respErr); err != nil {
			return err
		}
		return &respErr
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Printf("content = %+v\n", string(content))
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
	dec := NewDecoder(resp.Body)

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

// Login attempts to login a user given its username and password.
// If login succeeds, a session will be created.
// The account will be authentified and ready to perform requests
// on behalf of that user.
func (acc *Account) Login(username, password string) error {
	if acc.salt == nil {
		acc.initKeys(crypto.Rand(32), password)
	}

	lc, err := acc.loginChallenge(username)
	if err != nil {
		return err
	}

	req := LoginRequest{
		Username:  username,
		Challenge: lc.Challenge,
		Host:      "api.etebase.com",
		Action:    "login",
	}
	buf, err := Marshal(req)
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
		if err := NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return err
		}
		return &errResp
	}

	var loginResponse LoginResponse
	if err := NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}

	acc.session = &loginResponse
	return nil
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
	buf, err := Marshal(req)
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
	if err := NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
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
	if err := NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	log.Printf("body = %+v\n", body)

	return err
}
