package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of HTTPClient
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) GET(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*http.Response, error) {
	args := m.Called(ctx, url, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) POST(
	ctx context.Context,
	url string,
	body any,
	headers map[string]string,
) (*http.Response, error) {
	args := m.Called(ctx, url, body, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) PUT(
	ctx context.Context,
	url string,
	body any,
	headers map[string]string,
) (*http.Response, error) {
	args := m.Called(ctx, url, body, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) DELETE(
	ctx context.Context,
	url string,
	headers map[string]string,
) (*http.Response, error) {
	args := m.Called(ctx, url, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// Helper function to create an HTTP response with a body
func NewHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

// Helper function to create an HTTP response with JSON body
func NewJSONResponse(statusCode int, jsonBody string) *http.Response {
	resp := NewHTTPResponse(statusCode, jsonBody)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}
