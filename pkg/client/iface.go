package client

import (
	"net/http"
)

// A WS client that is able to make requests
type WealthsSimpleClient interface {
	Do(*http.Request) (*http.Response, error)
}

type client struct {
	c *http.Client
}
