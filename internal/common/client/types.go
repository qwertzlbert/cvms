package client

import "context"

// abstract Client interface
type Client interface {
	SetEndpoint(endpoint string) Client
	GetEndpoint() (string, error)
	Get(context context.Context, uri string) ([]byte, error)
	Post(context context.Context, uri string, body ...[]byte) ([]byte, error)
}
