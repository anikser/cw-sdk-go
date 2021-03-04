/*
Package rest provides a client for using the Cryptowatch REST API.
*/
package rest // import "code.cryptowat.ch/cw-sdk-go/client/rest"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/config"
	"code.cryptowat.ch/cw-sdk-go/version"
)

type RESTClient struct {
	url        string
	apiKey     string
	baseURLStr string
}

type RESTClientParams struct {
	URL    string
	APIKey string
}

func NewRESTClient(params *RESTClientParams) *RESTClient {
	if params == nil {
		params = &RESTClientParams{}
	}

	paramsCopy := *params
	params = &paramsCopy

	cfg := config.Get()

	if params.APIKey == "" {
		params.APIKey = cfg.APIKey
	}

	if params.URL == "" {
		params.URL = cfg.RESTURL
	}

	return &RESTClient{
		url:        params.URL,
		apiKey:     params.APIKey,
		baseURLStr: params.URL,
	}
}

type request struct {
	endpoint string
	params   map[string]string
}

var APIErrorResponse = errors.New("API error response")

func (rc *RESTClient) do(req request) (json.RawMessage, error) {
	u, err := url.Parse(rc.url)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u.Path = req.endpoint

	if len(req.params) > 0 {
		p := url.Values{}
		for k, v := range req.params {
			p.Add(k, v)
		}
		u.RawQuery = p.Encode()
	}

	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if rc.apiKey != "" {
		r.Header.Set("X-CW-API-Key", rc.apiKey)
	}

	r.Header.Set("User-Agent", fmt.Sprintf("cw-sdk-go@%s", version.Version))

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	// First check to see if there was an error response
	var response struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := dec.Decode(&response); err != nil {
		return nil, errors.Trace(err)
	}
	if response.Error != "" {
		return nil, errors.Annotate(APIErrorResponse, response.Error)
	}

	return response.Result, nil
}
