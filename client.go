package etebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/vmihailenco/msgpack"
)

type ClientOptions struct {
	Host string
	Env  string
}

func (opts ClientOptions) baseUrl(name string) string {
	return fmt.Sprintf("https://%s/%s/%s/api/v1", opts.Host, opts.Env, name)
}

var DefaultClientOptions = ClientOptions{
	Host: "api.etebase.com",
	Env:  "developer",
}

type Client struct {
	baseUrl string
}

func NewClient(name string, opts ClientOptions) *Client {
	return &Client{
		baseUrl: opts.baseUrl(name),
	}
}

func (c *Client) url(path string) string {
	url := c.baseUrl + path
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

func (c *Client) Post(path string, v interface{}) (*http.Response, error) {
	log.Printf("POST %s", c.url(path))
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	return http.Post(c.url(path), "application/msgpack", bytes.NewBuffer(body))
}

type Account struct {
	client   *Client
	password string

	salt                []byte
	mainKey, accountKey []byte
	authPub, authPriv   []byte
	idPub, idPriv       []byte
}

func NewAccount(c *Client) *Account {
	acc := &Account{
		client: c,
	}

	return acc
}

func (acc *Account) initKeys(password string) {
	acc.password = password

	acc.salt = Rand(32)
	acc.mainKey = DeriveKey(acc.salt, password)
	acc.accountKey = Rand(32)
	acc.authPub, acc.authPriv = GenrateKeyPair(acc.mainKey)
	acc.idPub, acc.idPriv = GenrateKeyPair(Rand(32))
}

func (acc *Account) Signup(user User, password string) error {
	acc.initKeys(password)

	encrypedContent, err := Encrypt(acc.mainKey, append(acc.accountKey, acc.idPriv...))
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
		if err := json.NewDecoder(resp.Body).Decode(&respErr); err != nil {
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
	dec := json.NewDecoder(resp.Body)

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

func (acc *Account) Login(username, password string) (*LoginResponse, error) {
	if acc.password == "" {
		acc.initKeys(password)
	}

	challenge, err := acc.loginChallenge(username)
	if err != nil {
		return nil, err
	}

	buf, err := msgpack.Marshal(&LoginRequest{
		Username:  username,
		Challenge: challenge.Challenge,
		Host:      "api.etebase.com",
		Action:    "login",
	})
	if err != nil {
		return nil, err
	}

	//	mainKey := DeriveKey(challenge.Salt, password)
	sig := Sign(acc.authPriv, buf)
	resp, err := acc.client.Post("/authentication/login", struct {
		Response  []byte `msgpack:"response"`
		Signature []byte `msgpack:"signature"`
	}{buf, sig})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var loginResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return nil, err
	}

	return &loginResponse, nil
}
