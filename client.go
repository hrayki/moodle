package moodle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const MobileService = "moodle_mobile_app"

type Client struct {
	Site  string
	Token string
	HTTP  *http.Client
	Agent string
}

func New(site, token string) *Client {
	return &Client{Site: strings.TrimRight(site, "/"), Token: token}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return defaultHTTP
}

var defaultHTTP = &http.Client{Timeout: 60 * time.Second}

func (c *Client) post(ctx context.Context, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Site+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	req.Header.Set("Accept", "application/json")
	if c.Agent != "" {
		req.Header.Set("User-Agent", c.Agent)
	}
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, &HTTPError{Status: res.StatusCode, Body: clip(string(body), 500)}
	}
	return body, nil
}

func (c *Client) Call(ctx context.Context, function string, params any, out any) error {
	form, err := Values(params)
	if err != nil {
		return err
	}
	form.Set("wstoken", c.Token)
	form.Set("wsfunction", function)
	form.Set("moodlewsrestformat", "json")
	body, err := c.post(ctx, "/webservice/rest/server.php", form)
	if err != nil {
		return err
	}
	if err := checkError(body); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("moodle: %s: %w", function, err)
	}
	return nil
}

func (c *Client) CallRaw(ctx context.Context, function string, params any) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.Call(ctx, function, params, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
