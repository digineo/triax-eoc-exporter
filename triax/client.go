package triax

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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
		Timeout: time.Second * 10,
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

func (c *Client) login(ctx context.Context) error {
	HTTPClient.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   sessionCookieName,
		MaxAge: -1,
	}})

	req := loginRequest{Username: c.username, Password: c.password}
	res := loginResponse{}
	err := c.apiRequest(ctx, http.MethodPost, loginPath, &req, &res)

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

// apiRequest sends an API request to the controller. The path is constructed
// from c.base + "/api/" + path. The request parameter, if not nil, will be
// JSON encoded, and the JSON response is decoded into the response parameter.
//
//	req, res := requestType{...}, responseType{...}
//	err := c.apiRequest(ctx, "POST", "node/status", &req, &res)
func (c *Client) apiRequest(ctx context.Context, method, path string, request, response interface{}) error {
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

	err = json.Unmarshal(jsonData, &response)
	if Verbose || err != nil {
		log.Println(string(jsonData))
	}
	if err != nil {
		return fmt.Errorf("decoding response failed: %w", err)
	}

	return nil
}

func (c *Client) apiGet(ctx context.Context, path string, res interface{}) error {
	return c.apiRequest(ctx, http.MethodGet, path, nil, res)
}

func (c *Client) Metrics(ctx context.Context) (*Metrics, error) {
	eoc := syseocResponse{}       // EoC port names
	sys := sysinfoResponse{}      // uptime, memory
	ghn := ghnStatusResponse{}    // G.HN port status
	nodes := nodeStatusResponse{} // data for each AP

retry:
	retried := false
	if err := c.apiGet(ctx, sysinfoPath, &sys); err != nil {
		if errStatus, ok := err.(*ErrUnexpectedStatus); ok && errStatus.Status == http.StatusUnauthorized && !retried {
			err = c.login(ctx)
			if err == nil {
				retried = true
				goto retry
			}
		}

		return nil, err
	}

	if err := c.apiGet(ctx, syseocPath, &eoc); err != nil {
		return nil, err
	}

	if err := c.apiGet(ctx, ghnStatusPath, &ghn); err != nil {
		return nil, err
	}

	if err := c.apiGet(ctx, nodeStatusPath, &nodes); err != nil {
		return nil, err
	}

	m := &Metrics{}
	m.Uptime = sys.Uptime
	m.Load = sys.Load
	m.Memory.Total = sys.Memory.Total
	m.Memory.Free = sys.Memory.Free
	m.Memory.Shared = sys.Memory.Shared
	m.Memory.Buffered = sys.Memory.Buffered

	m.GhnPorts = make(map[string]*GhnPort)
	for _, port := range ghn {
		m.GhnPorts[strings.ToLower(port.Mac)] = &GhnPort{
			Number:              -1, // determined in next step
			EndpointsOnline:     port.Connected,
			EndpointsRegistered: port.Registered,
		}
	}

	for mac := range m.GhnPorts {
		if i := eoc.MacAddr.Index(mac); i >= 0 {
			m.GhnPorts[mac].Number = i + 1 // yep.
		}
	}

	m.Endpoints = make([]EndpointMetrics, len(nodes))
	i := 0
	for _, node := range nodes {
		ep := &m.Endpoints[i]
		ep.Name = node.Name
		ep.MAC = node.Mac
		ep.Status = node.Statusid
		ep.StatusText = node.Status
		ep.Uptime = node.Sysinfo.Uptime
		ep.Load = node.Sysinfo.Load
		ep.GhnPortNumber = -1
		ep.GhnStats = node.GhnStats
		ep.Statistics = node.Statistics

		if mac := node.GhnMaster; mac != "" {
			ep.GhnPortMac = mac
			ep.GhnPortNumber = eoc.MacAddr.Index(mac)
		}

		i++
	}

	return m, nil
}
