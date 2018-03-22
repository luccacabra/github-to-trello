/* Thin wrapper around github.com/luccacabra/trello */

package trello

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/levigross/grequests"
	"github.com/luccacabra/trello"
	"github.com/pkg/errors"
)

type ClientConfig struct {
	BoardName     string
	LabelCardName string
	LabelMap      map[string]string
}

type Client struct {
	config ClientConfig

	client *trello.Client
	board  *trello.Board

	labelIDMap map[string]string // label Name -> label ID
	listIDMap  map[string]string // list Name  -> *trello.List

}

func NewClient(key, token string, config ClientConfig) *Client {
	c := &Client{
		client: trello.NewClient(key, token),
		config: config,
	}

	c.labelIDMap = map[string]string{}
	c.listIDMap = map[string]string{}

	c.loadResources(config)

	return c
}

func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	if target != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Unexpected HTTP response code %s", resp.StatusCode))
		}

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(target); err != nil {
			return errors.Wrap(err, "JSON decode failed")
		}
	}
	return nil
}

func (c *Client) Delete(path string, params map[string]string, target interface{}) error {
	params["key"] = c.client.Key
	params["token"] = c.client.Token

	return c.client.Delete(path, params, target)
}

func (c *Client) Get(path string, params map[string]string, target interface{}) error {
	params["key"] = c.client.Key
	params["token"] = c.client.Token

	return c.client.Get(path, params, target)
}

func (c *Client) Post(path string, data map[string]string, target interface{}) error {
	// Trello prohibits more than 10 seconds/second per token
	c.client.Throttle()

	params := map[string]string{
		"key":   c.client.Key,
		"token": c.client.Token,
	}

	url := fmt.Sprintf("%s/%s", c.client.BaseURL, path)

	fmt.Printf("POST URL: %s\n", url)

	resp, err := grequests.Post(
		url,
		&grequests.RequestOptions{
			Data:   data,
			Params: params,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "HTTP POST failure on %s", url)
	}

	return c.parseResponse(resp.RawResponse, target)
}

func (c *Client) Put(path string, data map[string]string, target interface{}) error {
	// Trello prohibits more than 10 seconds/second per token
	c.client.Throttle()

	params := map[string]string{
		"key":   c.client.Key,
		"token": c.client.Token,
	}

	url := fmt.Sprintf("%s/%s", c.client.BaseURL, path)

	fmt.Printf("PUT URL: %s\n", url)

	resp, err := grequests.Put(
		url,
		&grequests.RequestOptions{
			Data:   data,
			Params: params,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "HTTP PUT failure on %s", url)
	}

	return c.parseResponse(resp.RawResponse, target)
}
