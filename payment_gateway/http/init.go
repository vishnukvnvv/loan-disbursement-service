package http

import (
	"context"
	"net/http"
	"time"
)

type HTTPClient interface {
	GET(ctx context.Context, url string, headers map[string]string) (*http.Response, error)
	POST(
		ctx context.Context,
		url string,
		body any,
		headers map[string]string,
	) (*http.Response, error)
	PUT(
		ctx context.Context,
		url string,
		body any,
		headers map[string]string,
	) (*http.Response, error)
	DELETE(ctx context.Context, url string, headers map[string]string) (*http.Response, error)
}

func NewHTTPClient() HTTPClient {
	return NewNetHTTPClient(30 * time.Second)
}
