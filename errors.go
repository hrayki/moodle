package moodle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidToken    = errors.New("moodle: the token is invalid or expired")
	ErrPassport        = errors.New("moodle: the sign-in link doesn't belong to this sign-in")
	ErrNoMobileService = errors.New("moodle: this site doesn't allow app access")
	ErrBrowserLogin    = errors.New("moodle: this site signs in through the browser")
)

type Error struct {
	Code      string `json:"errorcode"`
	Message   string `json:"message"`
	Exception string `json:"exception"`
	Debug     string `json:"debuginfo"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return "moodle: " + e.Message
	}
	return "moodle: " + e.Code
}

func (e *Error) Is(target error) bool {
	switch target {
	case ErrInvalidToken:
		return e.Code == "invalidtoken" || e.Code == "accessexception" && e.Message == "Invalid token - token expired"
	case ErrNoMobileService:
		return e.Code == "enablewsdescription" || e.Code == "servicenotavailable"
	}
	return false
}

type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("moodle: the site answered %d", e.Status)
}

func checkError(body []byte) error {
	body = bytes.TrimSpace(body)
	if len(body) == 0 || body[0] != '{' {
		return nil
	}
	var e struct {
		Error
		Err string `json:"error"`
	}
	if json.Unmarshal(body, &e) != nil {
		return nil
	}
	if e.Exception == "" && e.Code == "" && e.Err == "" {
		return nil
	}
	if e.Message == "" {
		e.Message = e.Err
	}
	return &e.Error
}
