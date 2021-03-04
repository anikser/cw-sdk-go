package rest

import (
	"testing"

	"code.cryptowat.ch/cw-sdk-go/common"
)

var orderBookResponse = `
{
  "result": {
    "asks": [
      [ 9196.6, 7 ],
      [ 9198.7, 1 ],
      [ 9198.8, 1.988183 ]
    ],
    "bids": [
      [ 9196.5, 0.009946 ],
      [ 9192.2, 0.04168 ]
    ],
    "seqNum": 508834
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`

func TestGetOrderBook(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/kraken/btcusd/orderbook"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of orderbook", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetOrderBook("kraken", "btcusd")
			},
			resp: orderBookResponse,
			wantResult: common.OrderBookSnapshot{
				SeqNum: 508834,

				Asks: []common.PublicOrder{
					common.PublicOrder{Price: dfs("9196.6"), Amount: dfs("7")},
					common.PublicOrder{Price: dfs("9198.7"), Amount: dfs("1")},
					common.PublicOrder{Price: dfs("9198.8"), Amount: dfs("1.988183")},
				},
				Bids: []common.PublicOrder{
					common.PublicOrder{Price: dfs("9196.5"), Amount: dfs("0.009946")},
					common.PublicOrder{Price: dfs("9192.2"), Amount: dfs("0.04168")},
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetOrderBookByID(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/v2/markets/1/book"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of orderbook", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetOrderBookByID(common.MarketID(1))
			},
			resp: orderBookResponse,
			wantResult: common.OrderBookSnapshot{
				SeqNum: 508834,

				Asks: []common.PublicOrder{
					common.PublicOrder{Price: dfs("9196.6"), Amount: dfs("7")},
					common.PublicOrder{Price: dfs("9198.7"), Amount: dfs("1")},
					common.PublicOrder{Price: dfs("9198.8"), Amount: dfs("1.988183")},
				},
				Bids: []common.PublicOrder{
					common.PublicOrder{Price: dfs("9196.5"), Amount: dfs("0.009946")},
					common.PublicOrder{Price: dfs("9192.2"), Amount: dfs("0.04168")},
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
