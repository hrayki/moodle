package moodle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	LoginInApp      = 1
	LoginInBrowser  = 2
	LoginInEmbedded = 3
)

type Provider struct {
	Name    string `json:"name"`
	IconURL string `json:"iconurl"`
	URL     string `json:"url"`
}

type PublicConfig struct {
	WWWRoot          string     `json:"wwwroot"`
	HTTPSWWWRoot     string     `json:"httpswwwroot"`
	SiteName         string     `json:"sitename"`
	TypeOfLogin      int        `json:"typeoflogin"`
	LaunchURL        string     `json:"launchurl"`
	WebServices      int        `json:"enablewebservices"`
	MobileService    int        `json:"enablemobilewebservice"`
	Maintenance      int        `json:"maintenanceenabled"`
	MaintenanceText  string     `json:"maintenancemessage"`
	ShowLoginForm    *int       `json:"showloginform"`
	AuthInstructions string     `json:"authinstructions"`
	ForgotPassword   string     `json:"forgottenpasswordurl"`
	LogoURL          string     `json:"logourl"`
	CompactLogoURL   string     `json:"compactlogourl"`
	Providers        []Provider `json:"identityproviders"`
	DisabledFeatures string     `json:"tool_mobile_disabledfeatures"`
	MFA              bool       `json:"tool_mfa_enabled"`
	Recaptcha        bool       `json:"enableloginrecaptcha"`
}

func (p *PublicConfig) Root() string {
	if p.HTTPSWWWRoot != "" {
		return strings.TrimRight(p.HTTPSWWWRoot, "/")
	}
	return strings.TrimRight(p.WWWRoot, "/")
}

func (p *PublicConfig) AppAccess() bool {
	return p.WebServices == 1 && p.MobileService == 1
}

func (p *PublicConfig) PasswordLogin() bool {
	if p.TypeOfLogin != LoginInApp {
		return false
	}
	return p.ShowLoginForm == nil || *p.ShowLoginForm == 1
}

func (p *PublicConfig) BrowserLogin() bool {
	return p.TypeOfLogin != LoginInApp || len(p.Providers) > 0
}

type Site struct {
	URL    string
	Config PublicConfig
	HTTP   *http.Client
}

func Discover(ctx context.Context, address string, hc *http.Client) (*Site, error) {
	if hc == nil {
		hc = defaultHTTP
	}
	candidates, err := siteCandidates(address)
	if err != nil {
		return nil, err
	}
	var first error
	for _, base := range candidates {
		cfg, err := publicConfig(ctx, hc, base)
		if err == nil {
			root := cfg.Root()
			if root == "" {
				root = base
			}
			return &Site{URL: root, Config: *cfg, HTTP: hc}, nil
		}
		if first == nil || errors.Is(err, ErrNoMobileService) {
			first = err
		}
	}
	return nil, first
}

func siteCandidates(address string) ([]string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, fmt.Errorf("moodle: give the site's address")
	}
	if !strings.Contains(address, "://") {
		address = "https://" + address
	}
	u, err := url.Parse(address)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, fmt.Errorf("moodle: %q isn't a web address", address)
	}
	u.RawQuery, u.Fragment = "", ""
	path := strings.TrimRight(u.Path, "/")
	for _, tail := range []string{"/login/index.php", "/my/index.php", "/my", "/index.php", "/course/view.php", "/login"} {
		if strings.HasSuffix(path, tail) {
			path = strings.TrimSuffix(path, tail)
			break
		}
	}
	u.Path = path
	base := strings.TrimRight(u.String(), "/")
	out := []string{base}
	if path == "" {
		out = append(out, base+"/moodle")
	}
	return out, nil
}

func publicConfig(ctx context.Context, hc *http.Client, base string) (*PublicConfig, error) {
	c := &Client{Site: base, HTTP: hc}
	body, err := json.Marshal([]map[string]any{{"index": 0, "methodname": "tool_mobile_get_public_config", "args": map[string]any{}}})
	if err != nil {
		return nil, err
	}
	raw, err := c.ajax(ctx, "/lib/ajax/service-nologin.php?info=tool_mobile_get_public_config", body)
	if err != nil {
		return nil, err
	}
	var out []struct {
		Error     bool            `json:"error"`
		Exception *Error          `json:"exception"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out) == 0 {
		if e := checkError(raw); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("moodle: %s doesn't look like a Moodle site", base)
	}
	if out[0].Error {
		if out[0].Exception != nil {
			return nil, out[0].Exception
		}
		return nil, ErrNoMobileService
	}
	var cfg PublicConfig
	if err := json.Unmarshal(out[0].Data, &cfg); err != nil {
		return nil, err
	}
	if !cfg.AppAccess() {
		return &cfg, ErrNoMobileService
	}
	return &cfg, nil
}

func (c *Client) ajax(ctx context.Context, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Site+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}
