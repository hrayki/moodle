package moodle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
)

func Values(params any) (url.Values, error) {
	out := url.Values{}
	if params == nil {
		return out, nil
	}
	var tree any
	switch p := params.(type) {
	case url.Values:
		for k, v := range p {
			out[k] = append([]string(nil), v...)
		}
		return out, nil
	case json.RawMessage:
		if err := decode(p, &tree); err != nil {
			return nil, err
		}
	default:
		raw, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		if err := decode(raw, &tree); err != nil {
			return nil, err
		}
	}
	m, ok := tree.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("moodle: parameters must be an object")
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flatten(out, k, m[k])
	}
	return out, nil
}

func decode(raw []byte, v *any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	return d.Decode(v)
}

func flatten(out url.Values, key string, v any) {
	switch x := v.(type) {
	case nil:
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			flatten(out, key+"["+k+"]", x[k])
		}
	case []any:
		for i, e := range x {
			flatten(out, key+"["+strconv.Itoa(i)+"]", e)
		}
	case bool:
		if x {
			out.Set(key, "1")
		} else {
			out.Set(key, "0")
		}
	case json.Number:
		out.Set(key, x.String())
	case string:
		out.Set(key, x)
	default:
		out.Set(key, fmt.Sprint(x))
	}
}
