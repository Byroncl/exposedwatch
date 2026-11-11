// Package dockerapi lists published container ports through the Docker
// Engine API served on a unix socket, using only the standard library.
package dockerapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// Client queries the Docker Engine API.
type Client struct {
	httpClient *http.Client
}

// New returns a Client for the given socket. An empty path means
// DefaultSocketPath.
func New(socketPath SocketPath) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	dialer := &net.Dialer{}
	return &Client{httpClient: &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, unixNetwork, string(socketPath))
			},
		},
	}}
}

// listContainers fetches and decodes the container list.
func (client *Client) listContainers(ctx context.Context) ([]container, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, containersURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.httpClient.Do(request)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		var operationError *net.OpError
		if errors.As(err, &operationError) && operationError.Op == dialOperation {
			return nil, fmt.Errorf(unavailableMessage, ErrDockerUnavailable, err)
		}
		return nil, fmt.Errorf(requestFailedMessage, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(unexpectedStatusFmt, response.Status)
	}

	var containers []container
	if err := json.NewDecoder(response.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf(decodeResponseMessage, err)
	}
	return containers, nil
}
