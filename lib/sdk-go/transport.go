package sdk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ServiceConfig struct {
	Slug    string
	BaseURL string
	Token   string
}

type RoundTripper func(method, rawURL string, headers map[string]string, body *string) (status int, respBody string, err error)

type Executor struct {
	trip RoundTripper
}

func defaultTrip(method, rawURL string, headers map[string]string, body *string) (int, string, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewBufferString(*body)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return 0, "", transportErr(err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: time.Duration(TimeoutSeconds) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, "", transportErr(err.Error())
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, "", transportErr(err.Error())
	}
	return res.StatusCode, string(raw), nil
}

func makeService(slug string, row ServiceOptions) ServiceConfig {
	base := strings.TrimRight(row.BaseURL, "/")
	if base == "" {
		base = DefaultBaseURL(slug)
	}
	return ServiceConfig{Slug: slug, BaseURL: base, Token: strings.TrimSpace(row.Token)}
}

func (e *Executor) Call(service ServiceConfig, method, path string, query map[string]string, jsonBody any, auth bool) (map[string]any, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, clientErr("Path must start with /.", "")
	}
	rawURL := service.BaseURL + path
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			if v == "" {
				continue
			}
			q.Set(k, v)
		}
		enc := q.Encode()
		if enc != "" {
			if strings.Contains(rawURL, "?") {
				rawURL += "&" + enc
			} else {
				rawURL += "?" + enc
			}
		}
	}
	var body *string
	headers := map[string]string{"Accept": "application/json"}
	if jsonBody != nil {
		encoded, err := json.Marshal(jsonBody)
		if err != nil {
			return nil, transportErr(err.Error())
		}
		if len(encoded) > MaxBodyBytes {
			return nil, transportErr("Request body exceeds 65536 bytes.")
		}
		s := string(encoded)
		body = &s
		headers["Content-Type"] = "application/json"
	}
	if auth {
		if service.Token == "" {
			return nil, authErr("API token is required.", "")
		}
		headers = AuthHeaders(service.Token, jsonBody != nil)
	}
	trip := e.trip
	if trip == nil {
		trip = defaultTrip
	}
	status, respBody, err := trip(method, rawURL, headers, body)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if respBody != "" {
		_ = json.Unmarshal([]byte(respBody), &payload)
	}
	errText := ""
	if v, ok := payload["error"].(string); ok {
		errText = v
	}
	if status == 401 {
		msg := errText
		if msg == "" {
			msg = "Unauthorized."
		}
		return nil, authErr(msg, errText)
	}
	if status == 400 {
		msg := errText
		if msg == "" {
			msg = "Bad request."
		}
		return nil, clientErr(msg, errText)
	}
	if status >= 400 {
		msg := errText
		if msg == "" {
			msg = httpStatusMessage(status)
		}
		return nil, apiErr(status, msg, errText)
	}
	return payload, nil
}
