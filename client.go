package etebase

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/vmihailenco/msgpack"

	"github.com/etesync/etebase-go/internal/crypto"
)

var (
	ErrNoToken = errors.New("account has no token set, use Signup or Login first")
)

// ClientOptions allow you control specific options of the client.
// Most of the users should use DefaultClientOptions when constructing the
// client.
type ClientOptions struct {
	// Host is the Etebase server host.
	Host string
	// Env is the Etebase server environment.
	// Possible values are `partner` or `developer`.
	Env string
}

func (opts ClientOptions) baseUrl(name string) string {
	return fmt.Sprintf("https://%s/%s/%s/api/v1", opts.Host, opts.Env, name)
}

// DefaultClientOptions will make your client point to the official Etebase
// server in 'developer' mode.
var DefaultClientOptions = ClientOptions{
	Host: "api.etebase.com",
	Env:  "developer",
}

// Client implements the network client to use to interact with the Etebase
// server.
type Client struct {
	baseUrl string
	token   string
}

// NewClient returns a new client object given a name (your etebase account name),
// and options inside the ClientOptions struct.
func NewClient(name string, opts ClientOptions) *Client {
	return &Client{
		baseUrl: opts.baseUrl(name),
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

// Account represents a user account and is the main object for all user
// interactions and data manipulation.
type Account struct {
	client *Client
	token  string

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

func (acc *Account) initKeys(password string) {
	acc.salt = crypto.Rand(32)
	acc.mainKey = crypto.DeriveKey(acc.salt, password)
	acc.accountKey = crypto.Rand(32)
	acc.authPub, acc.authPriv = crypto.GenrateKeyPair(acc.mainKey)
	acc.idPub, acc.idPriv = crypto.GenrateKeyPair(crypto.Rand(32))
}

// Signup attempts to signup a new user into the server given a user and a
// password.
func (acc *Account) Signup(user User, password string) error {
	if acc.salt == nil {
		acc.initKeys(password)
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
// If login succeeds account will be authentified and ready to perform requests
// on behalf of that user.
func (acc *Account) Login(username, password string) error {
	if acc.salt == nil {
		acc.initKeys(password)
	}

	challenge, err := acc.loginChallenge(username)
	if err != nil {
		return err
	}

	buf, err := Marshal(&LoginRequest{
		Username:  username,
		Challenge: challenge.Challenge,
		Host:      "api.etebase.com",
		Action:    "login",
	})
	if err != nil {
		return err
	}

	//	mainKey := DeriveKey(challenge.Salt, password)
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
	if err := NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return err
	}
	acc.token = loginResponse.Token

	return nil
}

// Collection is not implemented yet.
func (acc *Account) Collection() error {
	if acc.token == "" {
		return ErrNoToken
	}

	resp, err := acc.client.WithToken(acc.token).Post("/collection/", nil)
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
