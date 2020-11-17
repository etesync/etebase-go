package etebase

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/vmihailenco/msgpack"
)

type Client struct {
	baseUrl string
}

func NewClient(url string) *Client {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimRight(url, "/")
	}
	return &Client{
		baseUrl: url,
	}
}

func (c *Client) url(p string) string {
	url := c.baseUrl + "/api/v1" + p
	if !strings.HasSuffix(url, "/") {
		return url + "/"
	}
	return url
}

func (c *Client) post(path string, v interface{}) (*http.Response, error) {
	log.Printf("POST %s", c.url(path))
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	return http.Post(c.url(path), "application/msgpack", bytes.NewBuffer(body))
}

func (c *Client) Signup(username, password string) error {
	var (
		salt          = Rand(32)
		mainKey       = DeriveKey(salt, password)
		accountKey    = Rand(32)
		authPub, _    = GenrateKeyPair(mainKey)
		idPub, idPriv = GenrateKeyPair(Rand(32))
	)

	encryptedContent, err := Encrypt(mainKey, append(accountKey, idPriv...))
	if err != nil {
		return err
	}

	body := Signup{
		User: User{
			Username: username,
			Email:    "gchain@pm.me",
		},
		Salt:             salt,
		LoginPubkey:      authPub,
		PubKey:           idPub,
		EncryptedContent: encryptedContent,
	}

	resp, err := c.post("/authentication/signup", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var vErr Error
		if err := json.NewDecoder(resp.Body).Decode(&vErr); err != nil {
			return err
		}
		return &vErr
	}

	return nil
}
