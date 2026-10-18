package moodle

import (
	"bytes"
	"strconv"
	"strings"
	"time"
)

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(bytes.TrimSpace(data)), `"`)
	switch strings.ToLower(s) {
	case "", "0", "false", "null":
		*b = false
	default:
		*b = true
	}
	return nil
}

type Int int64

func (n *Int) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(bytes.TrimSpace(data)), `"`)
	if s == "" || s == "null" {
		*n = 0
		return nil
	}
	if s == "true" {
		*n = 1
		return nil
	}
	if s == "false" {
		*n = 0
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*n = Int(v)
	return nil
}

type Time int64

func (t Time) Time() time.Time {
	if t <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(t), 0)
}

func (t Time) IsZero() bool { return t <= 0 }

func (t *Time) UnmarshalJSON(data []byte) error {
	var n Int
	if err := n.UnmarshalJSON(data); err != nil {
		return err
	}
	*t = Time(n)
	return nil
}

type Float float64

func (f *Float) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(bytes.TrimSpace(data)), `"`)
	if s == "" || s == "null" || s == "-" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		*f = 0
		return nil
	}
	*f = Float(v)
	return nil
}

type Warning struct {
	Item        string `json:"item"`
	ItemID      Int    `json:"itemid"`
	WarningCode string `json:"warningcode"`
	Message     string `json:"message"`
}

type File struct {
	Name         string `json:"filename"`
	Path         string `json:"filepath"`
	Size         Int    `json:"filesize"`
	URL          string `json:"fileurl"`
	TimeModified Time   `json:"timemodified"`
	MimeType     string `json:"mimetype"`
	Type         string `json:"type,omitempty"`
	Author       string `json:"author,omitempty"`
}
