package rest

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"
)

type priceServer struct {
	Price decimal.Decimal `json:"price"`
}

// GetPrice returns latest price on the given market.
func (c *RESTClient) GetPrice(
	exchangeSymbol string, pairSymbol string,
) (decimal.Decimal, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s/price", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return decimal.Decimal{}, errors.Trace(err)
	}

	ret := priceServer{}
	err = json.Unmarshal(result, &ret)

	return ret.Price, err
}
