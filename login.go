package moodle

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

type Token struct {
	Token   string `json:"token"`
	Private string `json:"privatetoken,omitempty"`
}

func (s *Site) Client(t *Token) *Client {
	c := New(s.URL, "")
	c.HTTP = s.HTTP
	if t != nil {
		c.Token = t.Token
	}
	return c
}

func (s *Site) Login(ctx context.Context, username, password string) (*Token, error) {
	form := url.Values{"username": {username}, "password": {password}, "service": {MobileService}}
	body, err := s.Client(nil).post(ctx, "/login/token.php", form)
	if err != nil {
		return nil, err
	}
	if err := checkError(body); err != nil {
		return nil, err
	}
	var t Token
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	if t.Token == "" {
		return nil, fmt.Errorf("moodle: the site gave no token")
	}
	return &t, nil
}

type LaunchOptions struct {
	Scheme   string
	Confirm  bool
	Provider *Provider
}

type Launch struct {
	URL      string   `json:"url"`
	Passport string   `json:"passport"`
	Roots    []string `json:"roots"`
}

func (s *Site) Launch(o LaunchOptions) (*Launch, error) {
	scheme := o.Scheme
	if scheme == "" {
		scheme = "moodlemobile"
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	passport := hex.EncodeToString(b)
	base := s.Config.LaunchURL
	if base == "" {
		base = s.URL + "/admin/tool/mobile/launch.php"
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("service", MobileService)
	q.Set("passport", passport)
	q.Set("urlscheme", scheme)
	if o.Confirm {
		q.Set("confirmed", "1")
	}
	if o.Provider != nil {
		if p, err := url.Parse(o.Provider.URL); err == nil && p.Query().Get("id") != "" {
			q.Set("oauthsso", p.Query().Get("id"))
		}
	}
	u.RawQuery = q.Encode()
	var roots []string
	for _, r := range []string{s.Config.WWWRoot, s.Config.HTTPSWWWRoot, s.URL} {
		if r != "" && !slices.Contains(roots, r) {
			roots = append(roots, r)
		}
	}
	return &Launch{URL: u.String(), Passport: passport, Roots: roots}, nil
}

func (l *Launch) Token(link string) (*Token, error) {
	link = strings.TrimSpace(link)
	if i := strings.Index(link, "token="); i >= 0 {
		link = link[i+len("token="):]
	}
	if v, err := url.QueryUnescape(link); err == nil {
		link = v
	}
	link = strings.TrimRight(strings.TrimSpace(link), "/#")
	raw, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		if raw, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(link, "=")); err != nil {
			return nil, fmt.Errorf("moodle: that isn't a sign-in link")
		}
	}
	parts := strings.Split(string(raw), ":::")
	if len(parts) < 2 || parts[1] == "" {
		return nil, fmt.Errorf("moodle: that isn't a sign-in link")
	}
	ok := false
	for _, r := range l.Roots {
		sum := md5.Sum([]byte(r + l.Passport))
		if hex.EncodeToString(sum[:]) == parts[0] {
			ok = true
		}
	}
	if !ok {
		return nil, ErrPassport
	}
	t := &Token{Token: parts[1]}
	if len(parts) > 2 {
		t.Private = parts[2]
	}
	return t, nil
}
