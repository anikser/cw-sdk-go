package rest

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
)

type orderbookServer struct {
	Asks   [][]decimal.Decimal `json:"asks"`
	Bids   [][]decimal.Decimal `json:"bids"`
	SeqNum common.SeqNum       `json:"seqNum"`
}

func (c *RESTClient) GetOrderBook(
	exchangeSymbol string, pairSymbol string,
) (common.OrderBookSnapshot, error) {
	resp, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s/orderbook", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return common.OrderBookSnapshot{}, errors.Trace(err)
	}

	return decodeOrderBookResponse(resp)
}

func (c *RESTClient) GetOrderBookByID(id common.MarketID) (
	common.OrderBookSnapshot, error,
) {
	resp, err := c.do(request{
		endpoint: fmt.Sprintf("v2/markets/%d/book", id),
	})
	if err != nil {
		return common.OrderBookSnapshot{}, errors.Trace(err)
	}

	return decodeOrderBookResponse(resp)
}

func decodeOrderBookResponse(in json.RawMessage) (common.OrderBookSnapshot, error) {
	var (
		ob  common.OrderBookSnapshot
		res orderbookServer
	)

	err := json.Unmarshal(in, &res)
	if err != nil {
		return common.OrderBookSnapshot{}, errors.Trace(err)
	}

	ob.SeqNum = res.SeqNum
	ob.Asks = make([]common.PublicOrder, 0, len(res.Asks))
	ob.Bids = make([]common.PublicOrder, 0, len(res.Bids))

	for i, v := range res.Asks {
		if len(v) != 2 {
			return common.OrderBookSnapshot{},
				errors.Errorf("Failed to decode order book response: ask #%d: expected tuple of 2 elements, got %v", i, v)
		}

		ob.Asks = append(ob.Asks, common.PublicOrder{
			Price:  v[0],
			Amount: v[1],
		})
	}

	for i, v := range res.Bids {
		if len(v) != 2 {
			return common.OrderBookSnapshot{},
				errors.Errorf("Failed to decode order book response: bid #%d: expected tuple of 2 elements, got %v", i, v)
		}

		ob.Bids = append(ob.Bids, common.PublicOrder{
			Price:  v[0],
			Amount: v[1],
		})
	}

	return ob, nil
}
