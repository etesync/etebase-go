package etebase

import (
	"bytes"
	"net/http"
	"path"
	"strings"

	"github.com/etesync/etebase-go/internal/codec"
	"github.com/vmihailenco/msgpack/v5"
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

	Logger interface {
		Logf(format string, v ...interface{})
	}
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
	opts    *ClientOptions
}

// NewClient returns a new client object given a name (your etebase account name),
// and options inside the ClientOptions struct.
func NewClient(opts ClientOptions) *Client {
	return &Client{
		baseUrl: opts.baseUrl(),
		opts:    &opts,
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

func (c *Client) log(format string, v ...interface{}) {
	if l := c.opts.Logger; l != nil {
		l.Logf(format, v...)
	}
}

// do sends a http Request with the right headers and verifies the status code
// before returning.
// If the status code isn't 200 <= x <= 400 it decodes the response error and
// closes the body.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	c.log("%-6s %s\n", req.Method, req.URL.Path)
	req.Header.Set("Content-Type", "application/msgpack")
	req.Header.Set("Accept", "application/msgpack")
	if t := c.token; t != "" {
		req.Header.Set("Authorization", "Token "+t)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	code := resp.StatusCode
	if code >= 200 && code <= 400 {
		return resp, nil
	}

	defer resp.Body.Close()

	var body ErrorResponse
	if err := codec.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return nil, &body
}

// Post posts an encoded value `v` to the server.
// `v` will be encoded using msgpack format.
func (c *Client) Post(path string, v interface{}) (*http.Response, error) {
	body, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url(path), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *Client) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.url(path), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) Host() string {
	return c.opts.Host
}
