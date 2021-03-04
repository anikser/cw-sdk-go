package rest

import (
	"encoding/json"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/common"
)

type ExchangeDescr struct {
	ID     int                 `json:"id"`
	Symbol string              `json:"symbol"`
	Name   string              `json:"name"`
	Active bool                `json:"active"`
	Routes ExchangeDescrRoutes `json:"routes"`
}

type ExchangeDescrRoutes struct {
	Markets string `json:"markets"`
}

type ExchangeDescrBrief struct {
	ID     int    `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Route  string `json:"route"`
	Active bool   `json:"active"`
}

type exchangesDescrServer struct {
	Result []ExchangeDescrBrief `json:"result"`
}

type exchangeDescrServer struct {
	Result ExchangeDescr `json:"result"`
}

// GetMarketBySymbol returns Market object based on exchange, base, and quote.
func (c *RESTClient) GetExchangeBySymbol(xchSymbol common.ExchangeSymbol) (
	common.Exchange, error,
) {
	result, err := c.do(request{
		endpoint: "v2/exchanges",
		params: map[string]string{
			"symbol": string(xchSymbol),
		},
	})
	if err != nil {
		return common.Exchange{}, errors.Trace(err)
	}

	exchange := common.Exchange{}
	err = json.Unmarshal(result, &exchange)

	return exchange, errors.Trace(err)
}

func (c *RESTClient) GetExchangesIndex() ([]ExchangeDescrBrief, error) {
	result, err := c.do(request{
		endpoint: "exchanges",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := []ExchangeDescrBrief{}
	err = json.Unmarshal(result, &ret)

	return ret, errors.Trace(err)
}

func (c *RESTClient) GetExchangeDescr(exchangeSymbol string) (*ExchangeDescr, error) {
	result, err := c.do(request{
		endpoint: "exchanges/" + exchangeSymbol,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := &ExchangeDescr{}
	err = json.Unmarshal(result, ret)

	return ret, errors.Trace(err)
}
