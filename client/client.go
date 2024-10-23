package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/digineo/triax-eoc-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
	endpoint *url.URL
	Username string
	Password string
	backend  types.Backend
}

var HTTPClient = http.Client{
	Timeout: time.Second * 30,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func init() {
	HTTPClient.Jar, _ = cookiejar.New(nil) // error is always nil
	HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

// NewClient creates a new Client instance. The URL must embed login
// credentials.
func NewClient(endpoint *url.URL) (*Client, error) {
	userinfo := endpoint.User
	if userinfo == nil {
		return nil, types.ErrMissingCredentials
	}

	pwd, _ := userinfo.Password()
	client := &Client{
		endpoint: endpoint,
		Username: userinfo.Username(),
		Password: pwd,
	}
	return client, nil
}

func (c *Client) Get(ctx context.Context, path string, res interface{}) error {
	return c.ApiRequest(ctx, http.MethodGet, path, nil, res)
}

const luaConfigPath = "/cgi.lua/config"

// GetConfig fetches the configuration from the controller
func (c *Client) GetConfig(ctx context.Context) (json.RawMessage, error) {
	msg := json.RawMessage{}
	err := c.Get(ctx, luaConfigPath, &msg)
	return msg, err
}

// SetConfig sets the configuration in the controller
func (c *Client) SetConfig(ctx context.Context, raw json.RawMessage) error {
	res := json.RawMessage{}
	err := c.ApiRequest(ctx, http.MethodPost, luaConfigPath, raw, res)
	return err
}

func (c *Client) SetCookie(nameAndValue string) {
	i := strings.Index(nameAndValue, "=")
	if i <= 0 {
		slog.Error("cannot split cookie", "value", nameAndValue)
		return
	}

	slog.Info("Set cookie",
		"host", c.endpoint.Host,
		"name", nameAndValue[:i],
		"value", nameAndValue[i+1:],
	)

	// Set cookie from response
	HTTPClient.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   nameAndValue[:i],
		Value:  nameAndValue[i+1:],
		MaxAge: 0,
	}})
}

func (c *Client) withBackend(ctx context.Context, f func(types.Backend) error) error {
	if c.backend == nil {
		err := c.setupBackend(ctx)
		if err != nil {
			return err
		}
	}

	return f(c.backend)
}

func (c *Client) setupBackend(ctx context.Context) error {
	c.backend = Try(ctx, c)
	if c.backend == nil {
		return errors.New("no usable backend found")
	}

	return nil
}

func (c *Client) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return c.withBackend(ctx, func(backend types.Backend) error {
		return backend.Collect(ctx, ch)
	})
}

// calls apiRequestRaw and does a login on unauthorized status
func (c *Client) ApiRequest(ctx context.Context, method, path string, request, response interface{}) error {
	return c.withBackend(ctx, func(backend types.Backend) error {

		retried := false
	retry:
		errStatus := &types.ErrUnexpectedStatus{}
		_, err := c.ApiRequestRaw(ctx, method, path, request, response)

		if errors.As(err, &errStatus) && errStatus.Status == http.StatusUnauthorized && !retried {

			if c.setupBackend(ctx) == nil {
				retried = true
				goto retry
			}
		}

		return err
	})
}

// apiRequestRaw sends an API request to the controller. The path is constructed
// from c.base + "/api/" + path. The request parameter, if not nil, will be
// JSON encoded, and the JSON response is decoded into the response parameter.
//
//	req, res := requestType{...}, responseType{...}
//	err := c.ApiRequestRaw(ctx, "POST", "node/status", &req, &res)
func (c *Client) ApiRequestRaw(ctx context.Context, method, path string, request, response interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s://%s/%s", c.endpoint.Scheme, c.endpoint.Host, strings.TrimPrefix(path, "/"))

	slog.Info("HTTP Request", "method", method, "url", url)

	var body io.Reader
	if request != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(request); err != nil {
			return nil, fmt.Errorf("encoding body failed: %w", err)
		}
		body = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("cannot construct request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(res.Body)

		return res, &types.ErrUnexpectedStatus{
			Method:   method,
			URL:      url,
			Location: res.Header.Get("Location"),
			Status:   res.StatusCode,
			Body:     data,
		}
	}

	jsonData, err := io.ReadAll(res.Body)
	if err != nil {
		return res, err
	}

	/*
		contentType := res.Header.Get("Content-Type")

		if contentType != "application/json" {
			slog.Error("unexpected content type", "contentType", contentType, "body", string(jsonData))
			return fmt.Errorf("unexpected content-type: %s", contentType)
		}
	*/

	if response != nil {
		err = json.Unmarshal(jsonData, &response)

		if err != nil {
			slog.Error("response received", "json", string(jsonData))
			return res, fmt.Errorf("decoding response failed: %w", err)
		}
		slog.Debug("response received", "json", string(jsonData))

	}

	return res, nil
}
