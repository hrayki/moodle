package moodle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Download struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
	Name        string
}

func (c *Client) FileURL(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	site, err := url.Parse(c.Site)
	if err != nil {
		return "", err
	}
	if u.Host != site.Host {
		return "", fmt.Errorf("moodle: %s isn't on this site", u.Host)
	}
	if strings.Contains(u.Path, "/pluginfile.php") && !strings.Contains(u.Path, "/webservice/pluginfile.php") {
		u.Path = strings.Replace(u.Path, "/pluginfile.php", "/webservice/pluginfile.php", 1)
	}
	q := u.Query()
	q.Set("token", c.Token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) Download(ctx context.Context, fileURL string) (*Download, error) {
	full, err := c.FileURL(fileURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return nil, err
	}
	if c.Agent != "" {
		req.Header.Set("User-Agent", c.Agent)
	}
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, &HTTPError{Status: res.StatusCode}
	}
	ct := res.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		defer res.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		if err := checkError(body); err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(body))
	}
	name := ""
	if u, err := url.Parse(fileURL); err == nil {
		parts := strings.Split(u.Path, "/")
		name, _ = url.PathUnescape(parts[len(parts)-1])
	}
	return &Download{Body: res.Body, ContentType: ct, Size: res.ContentLength, Name: name}, nil
}

type Uploaded struct {
	ItemID    Int    `json:"itemid"`
	Component string `json:"component"`
	FileArea  string `json:"filearea"`
	FileName  string `json:"filename"`
	FilePath  string `json:"filepath"`
	ContextID Int    `json:"contextid"`
	Error     string `json:"error"`
}

func (c *Client) Upload(ctx context.Context, itemID Int, name string, body io.Reader) (*Uploaded, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		err := func() error {
			for k, v := range map[string]string{"token": c.Token, "filearea": "draft", "itemid": fmt.Sprint(int64(itemID)), "filepath": "/"} {
				if err := mw.WriteField(k, v); err != nil {
					return err
				}
			}
			part, err := mw.CreateFormFile("file_1", name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(part, body); err != nil {
				return err
			}
			return mw.Close()
		}()
		pw.CloseWithError(err)
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Site+"/webservice/upload.php", pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if err := checkError(raw); err != nil {
		return nil, err
	}
	var out []Uploaded
	if err := json.Unmarshal(raw, &out); err != nil || len(out) == 0 {
		return nil, fmt.Errorf("moodle: the upload gave no file back")
	}
	if out[0].Error != "" {
		return nil, fmt.Errorf("moodle: %s", out[0].Error)
	}
	return &out[0], nil
}
