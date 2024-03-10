package zerion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type (
	Client struct {
		client  *http.Client
		apiURL  string
		authKey string
	}
)

func NewClient(apiURL, authKey string, client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}

	return &Client{
		client:  client,
		apiURL:  apiURL,
		authKey: authKey,
	}
}

// GetWalletPositions Get list of wallet's fungible positions
// see all parameters here: https://developers.zerion.io/reference/listwalletpositions
func (c *Client) GetWalletPositions(address string) (*WalletPositions, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("wallets/%s/positions/", address),
		"wallet-positions",
		map[string]string{
			"filter[positions]": "only_simple",
			"filter[trash]":     "only_non_trash",
			"currency":          "usd",
			"sort":              "value",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request do: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var positions WalletPositions
	if err = json.Unmarshal(body, &positions); err != nil {
		return nil, fmt.Errorf("unmarshal body: %w", err)
	}

	return &positions, nil
}

func (c *Client) buildRequest(method, subURL, alias string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", c.apiURL, subURL),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("alias", alias)

	return c.withAuth(req), nil
}

func (c *Client) withAuth(req *http.Request) *http.Request {
	if c.authKey == "" {
		return req
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.authKey))

	return req
}
