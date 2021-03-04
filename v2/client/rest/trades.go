package rest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
)

// GetTrades returns latest trades on the given market, sorted in ascending
// order.
func (c *RESTClient) GetTrades(
	exchangeSymbol string, pairSymbol string,
) ([]common.PublicTrade, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s/trades", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := [][]decimal.Decimal{}

	err = json.Unmarshal(result, &res)
	if err != nil {
		return nil, errors.Trace(err)
	}

	trades := make([]common.PublicTrade, 0, len(res))

	for _, ts := range res {
		if len(ts) != 4 {
			return nil, errors.Errorf("invalid trade format: %+v", ts)
		}

		trades = append(trades, common.PublicTrade{
			Timestamp: time.Unix(ts[1].IntPart(), 0),
			Price:     ts[2],
			Amount:    ts[3],
		})
	}

	return trades, nil
}
