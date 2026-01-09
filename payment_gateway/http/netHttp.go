package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type NetHTTPClient struct {
	client *http.Client
}

func NewNetHTTPClient(timeout time.Duration) *NetHTTPClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &NetHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *NetHTTPClient) GET(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, url, nil, headers)
}

func (c *NetHTTPClient) POST(
	ctx context.Context,
	url string,
	body any,
	headers map[string]string,
) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, url, body, headers)
}

func (c *NetHTTPClient) PUT(
	ctx context.Context,
	url string,
	body any,
	headers map[string]string,
) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, url, body, headers)
}

func (c *NetHTTPClient) DELETE(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, url, nil, headers)
}

func (c *NetHTTPClient) do(
	ctx context.Context,
	method, url string,
	body any,
	headers map[string]string,
) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			reader = bytes.NewBuffer(v)
		case string:
			reader = bytes.NewBufferString(v)
		default:
			buf, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewBuffer(buf)
			// default to JSON if content-type not explicitly set
			if headers == nil {
				headers = map[string]string{}
			}
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}
