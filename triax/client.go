package triax

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
)

type Client struct {
	MAC      string
	endpoint *url.URL
	base     string
	username string
	password string
	http     *http.Client
	cookies  http.CookieJar
}

// NewClient creates a new Client instance. The URL must embed login
// credentials.
func NewClient(url string, insecure bool, mac string) (*Client, error) {
	username, password, endpoint, err := extractCredentials(url)
	if err != nil {
		return nil, &ErrInvalidEndpoint{err}
	}

	jar, _ := cookiejar.New(nil) // error is always nil
	client := &Client{
		MAC:      mac,
		endpoint: endpoint,
		base:     endpoint.String(),
		username: username,
		password: password,
		http:     &http.Client{Jar: jar},
		cookies:  jar,
	}

	if insecure {
		// most installations will have a self-signed certificate
		client.http.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client, err
}

func extractCredentials(uri string) (user, pass string, endpoint *url.URL, err error) {
	endpoint, err = url.Parse(uri)
	if err != nil {
		return
	}

	pass, _ = endpoint.User.Password()
	user = endpoint.User.Username()
	if pass == "" || user == "" {
		err = ErrMissingCredentials
		return
	}

	endpoint.User = nil
	endpoint.Path = ""
	return
}

func (c *Client) Login(ctx context.Context) error {
	c.http.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   "sessionId",
		MaxAge: -1,
	}})

	req := loginRequest{Username: c.username, Password: c.password}
	res := loginResponse{}
	code, err := c.apiRequest(ctx, http.MethodPost, loginPath, &req, &res)

	if err != nil {
		return err
	} else if code != http.StatusOK {
		return &genericError{res.Message}
	}

	if !strings.HasPrefix(res.Cookie, "sessionId=") {
		return &genericError{fmt.Sprintf("unexpected cookie: %s", res.Cookie)}
	}

	c.http.Jar.SetCookies(c.endpoint, []*http.Cookie{{
		Name:   "sessionId",
		Value:  strings.TrimPrefix(res.Cookie, "sessionId="),
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
//	code, err := c.apiRequest(ctx, "POST", "node/status", &req, &res)
func (c *Client) apiRequest(ctx context.Context, method, path string, request, response interface{}) (int, error) {
	url := fmt.Sprintf("%s/api/%s", c.base, strings.TrimPrefix(path, "/"))

	var body io.Reader
	if request != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(request); err != nil {
			return 0, fmt.Errorf("encoding body failed: %w", err)
		}
		body = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, fmt.Errorf("cannot construct request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, res.Body)

	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		log.Println(buf.String())
		return res.StatusCode, fmt.Errorf("decoding response failed: %w", err)
	}
	return res.StatusCode, nil
}

func (c *Client) apiGet(ctx context.Context, path string, res interface{}) (int, error) {
	return c.apiRequest(ctx, http.MethodGet, path, nil, res)
}

func (c *Client) FetchData(ctx context.Context) (*Metrics, error) {
	eoc := syseocResponse{}       // EoC port names
	sys := sysinfoResponse{}      // uptime, memory
	ghn := ghnStatusResponse{}    // G.HN port status
	nodes := nodeStatusResponse{} // data for each AP

	results := map[string]struct {
		data interface{}
		err  error
	}{
		syseocPath:     {data: &eoc},
		sysinfoPath:    {data: &sys},
		ghnStatusPath:  {data: &ghn},
		nodeStatusPath: {data: &nodes},
	}

	wg := sync.WaitGroup{}
	wg.Add(5)

	for path := range results {
		// go func(path string) {
		res := results[path]
		_, res.err = c.apiGet(ctx, path, res.data)
		log.Printf("[%s] fetched %s: error = %v", c.MAC, path, res.err)
		wg.Done()
		// }(path)
	}

	wg.Wait()

	m := &Metrics{}
	if err := results[sysinfoPath].err; err != nil {
		return m, fmt.Errorf("sysinfo failed: %w", err)
	}

	m.Up = 1
	m.Uptime = sys.Uptime
	m.Load = sys.Load
	m.Memory.Total = sys.Memory.Total
	m.Memory.Free = sys.Memory.Free
	m.Memory.Shared = sys.Memory.Shared
	m.Memory.Buffered = sys.Memory.Buffered

	if err := results[ghnStatusPath].err; err != nil {
		return m, fmt.Errorf("ghn status failed: %w", err)
	}

	m.GhnPorts = make(map[string]*GhnPort)
	for _, port := range ghn {
		m.GhnPorts[strings.ToLower(port.Mac)] = &GhnPort{
			Number:              -1, // determined in next step
			EndpointsOnline:     port.Connected,
			EndpointsRegistered: port.Registered,
		}
	}

	if err := results[syseocPath].err; err != nil {
		return m, fmt.Errorf("eoc status failed: %w", err)
	}

	for mac := range m.GhnPorts {
		if i := eoc.MacAddr.Index(mac); i >= 0 {
			m.GhnPorts[mac].Number = i + 1 // yep.
		}
	}

	if err := results[nodeStatusPath].err; err != nil {
		return m, fmt.Errorf("nodes status failed: %w", err)
	}

	m.Endpoints = make([]*EndpointMetrics, 0, len(nodes))
	for _, node := range nodes {
		ep := &EndpointMetrics{
			Name:          node.Name,
			MAC:           node.Mac,
			Status:        node.Statusid,
			StatusText:    node.Status,
			Uptime:        node.Sysinfo.Uptime,
			Load:          node.Sysinfo.Load,
			GhnPortNumber: -1,
			Clients:       make(map[string]int),
		}
		if mac := node.GhnMaster; mac != "" {
			ep.GhnPortMac = mac
			ep.GhnPortNumber = eoc.MacAddr.Index(mac)
		}
		for _, client := range node.Clients {
			ep.Clients[client.Band]++
		}
	}

	return m, nil
}
