package client

import (
	"context"
	"net/http"
)

// Reload asks the running server to reread its own configuration files.
func (c *Client) Reload(ctx context.Context) error {
	return c.requestJSON(ctx, http.MethodPost, "/api/reload", nil, nil)
}
