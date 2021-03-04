package rest

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
)

type PairDescr struct {
	ID      int           `json:"id"`
	Symbol  string        `json:"symbol"`
	Route   string        `json:"route"`
	Base    PairSideDescr `json:"base"`
	Quote   PairSideDescr `json:"quote"`
	Markets []MarketDescr `json:"markets"`
}

type PairSideDescr struct {
	ID     int    `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Fiat   bool   `json:"fiat"`
	Route  string `json:"route"`
}

func (c *RESTClient) GetPairsIndex() ([]PairDescr, error) {
	result, err := c.do(request{
		endpoint: "pairs",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := []PairDescr{}
	err = json.Unmarshal(result, &res)

	return res, err
}

func (c *RESTClient) GetPairDescr(symbol string) (PairDescr, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("pairs/%s", symbol),
	})
	if err != nil {
		return PairDescr{}, errors.Trace(err)
	}

	res := PairDescr{}
	err = json.Unmarshal(result, &res)

	return res, err
}
