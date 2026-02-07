package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/peterhellberg/link"
)

func buildRequest(method, baseURL, path string, fields map[string]any, body io.Reader, headers []string) (*http.Request, error) {
	var u string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		u = path
	} else {
		var err error
		u, err = url.JoinPath(baseURL, path)
		if err != nil {
			return nil, fmt.Errorf("building URL: %w", err)
		}
	}

	if body == nil && len(fields) > 0 && (method == http.MethodGet || method == http.MethodHead || method == http.MethodDelete) {
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		q := parsed.Query()
		for k, v := range fields {
			q.Set(k, fmt.Sprint(v))
		}
		parsed.RawQuery = q.Encode()
		u = parsed.String()
		fields = nil
	}

	if body == nil && len(fields) > 0 {
		b, err := json.Marshal(fields)
		if err != nil {
			return nil, fmt.Errorf("encoding fields: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentTypeForPath(path))
	}

	for _, h := range headers {
		key, val, ok := strings.Cut(h, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q (must be key:value)", h)
		}
		req.Header.Set(strings.TrimSpace(key), strings.TrimSpace(val))
	}

	return req, nil
}

func nextPageURL(resp *http.Response) string {
	group := link.ParseResponse(resp)
	if next, ok := group["next"]; ok {
		return next.URI
	}
	return ""
}

func writeResponseHeaders(w io.Writer, resp *http.Response) {
	_, _ = fmt.Fprintf(w, "%s %s\n", resp.Proto, resp.Status)
	for key, vals := range resp.Header {
		for _, v := range vals {
			_, _ = fmt.Fprintf(w, "%s: %s\n", key, v)
		}
	}
	_, _ = fmt.Fprintln(w)
}
