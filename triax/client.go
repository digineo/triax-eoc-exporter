package triax

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	endpoint *url.URL
	username string
	password string
}

var (
	// Verbose increases verbosity.
	Verbose bool

	HTTPClient = http.Client{
		Timeout: time.Second * 30,
	}
)

func init() {
	HTTPClient.Jar, _ = cookiejar.New(nil) // error is always nil
	HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

const sessionCookieName = "sessionId"

// NewClient creates a new Client instance. The URL must embed login
// credentials.
func NewClient(endpoint *url.URL) (*Client, error) {
	userinfo := endpoint.User
	if userinfo == nil {
		return nil, ErrMissingCredentials
	}

	client := &Client{
		endpoint: endpoint,
		username: userinfo.Username(),
	}

	client.password, _ = userinfo.Password()

	return client, nil
}

func (c *Client) Get(ctx context.Context, path string, res interface{}) error {
	return c.apiRequest(ctx, http.MethodGet, path, nil, res)
}

func (c *Client) login(ctx context.Context) error {
	HTTPClient.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   sessionCookieName,
		MaxAge: -1,
	}})

	req := loginRequest{Username: c.username, Password: c.password}
	res := loginResponse{}
	err := c.apiRequestRaw(ctx, http.MethodPost, loginPath, &req, &res)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(res.Cookie, sessionCookieName+"=") {
		return &genericError{fmt.Sprintf("unexpected cookie: %s", res.Cookie)}
	}

	HTTPClient.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   sessionCookieName,
		Value:  strings.TrimPrefix(res.Cookie, sessionCookieName+"="),
		Raw:    res.Cookie,
		MaxAge: 0,
	}})

	return nil
}

// calls apiRequestRaw and does a login on unauthorized status
func (c *Client) apiRequest(ctx context.Context, method, path string, request, response interface{}) error {
	retried := false

retry:
	errStatus := &ErrUnexpectedStatus{}
	err := c.apiRequestRaw(ctx, method, path, request, response)

	if errors.As(err, &errStatus) && errStatus.Status == http.StatusUnauthorized && !retried {
		err = c.login(ctx)
		if err == nil {
			retried = true
			goto retry
		}
	}

	return err
}

// apiRequestRaw sends an API request to the controller. The path is constructed
// from c.base + "/api/" + path. The request parameter, if not nil, will be
// JSON encoded, and the JSON response is decoded into the response parameter.
//
//	req, res := requestType{...}, responseType{...}
//	err := c.apiRequestRaw(ctx, "POST", "node/status", &req, &res)
func (c *Client) apiRequestRaw(ctx context.Context, method, path string, request, response interface{}) error {
	url := fmt.Sprintf("%s://%s/api/%s", c.endpoint.Scheme, c.endpoint.Host, strings.TrimPrefix(path, "/"))

	if Verbose {
		log.Printf("%s %s", method, url)
	}

	var body io.Reader
	if request != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(request); err != nil {
			return fmt.Errorf("encoding body failed: %w", err)
		}
		body = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("cannot construct request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(res.Body)

		return &ErrUnexpectedStatus{
			Method: method,
			URL:    url,
			Status: res.StatusCode,
			Body:   data,
		}
	}

	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if response != nil {
		err = json.Unmarshal(jsonData, &response)
		if Verbose || err != nil {
			log.Println(string(jsonData))
		}
		if err != nil {
			return fmt.Errorf("decoding response failed: %w", err)
		}
	}

	return nil
}
