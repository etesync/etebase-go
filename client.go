package etebase

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
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
	log.Printf("GET  %s", path)
	return http.Get(c.url(path))
}
