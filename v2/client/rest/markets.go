package rest

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
)

type MarketDescr struct {
	ID       int    `json:"id"`
	Exchange string `json:"exchange"`
	Pair     string `json:"pair"`
	Active   bool   `json:"active"`
	Route    string `json:"route"`
}

func (c *RESTClient) GetMarketsIndex() ([]MarketDescr, error) {
	result, err := c.do(request{
		endpoint: "markets",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	market := []MarketDescr{}
	err = json.Unmarshal(result, &market)

	return market, err
}

func (c *RESTClient) GetMarketDescr(exchangeSymbol, pairSymbol string) (MarketDescr, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return MarketDescr{}, errors.Trace(err)
	}

	market := MarketDescr{}
	err = json.Unmarshal(result, &market)

	return market, err
}

func (c *RESTClient) GetExchangeMarketsDescr(exchangeSymbol string) ([]MarketDescr, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s", exchangeSymbol),
	})

	if err != nil {
		return nil, errors.Trace(err)
	}

	market := []MarketDescr{}
	err = json.Unmarshal(result, &market)

	return market, err
}
