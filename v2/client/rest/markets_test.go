package rest

import (
	"testing"
)

func TestGetMarketsIndex(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of markets index", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetMarketsIndex()
			},
			resp: `
{
  "result": [
    {
      "id": 1,
      "exchange": "bitfinex",
      "pair": "btcusd",
      "active": true,
      "route": "https://api.cryptowat.ch/markets/bitfinex/btcusd"
    },
    {
      "id": 2,
      "exchange": "bitfinex",
      "pair": "ltcusd",
      "active": true,
      "route": "https://api.cryptowat.ch/markets/bitfinex/ltcusd"
    }
  ],
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: []MarketDescr{
				MarketDescr{
					ID:       1,
					Exchange: "bitfinex",
					Pair:     "btcusd",
					Active:   true,
					Route:    "https://api.cryptowat.ch/markets/bitfinex/btcusd",
				},
				MarketDescr{
					ID:       2,
					Exchange: "bitfinex",
					Pair:     "ltcusd",
					Active:   true,
					Route:    "https://api.cryptowat.ch/markets/bitfinex/ltcusd",
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetMarketDescr(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/bitfinex/btcusd"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of market description", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetMarketDescr("bitfinex", "btcusd")
			},
			resp: `
{
  "result": {
    "id": 1,
    "exchange": "bitfinex",
    "pair": "btcusd",
    "active": true,
    "routes": {
      "price": "https://api.cryptowat.ch/markets/bitfinex/btcusd/price",
      "summary": "https://api.cryptowat.ch/markets/bitfinex/btcusd/summary",
      "orderbook": "https://api.cryptowat.ch/markets/bitfinex/btcusd/orderbook",
      "trades": "https://api.cryptowat.ch/markets/bitfinex/btcusd/trades",
      "ohlc": "https://api.cryptowat.ch/markets/bitfinex/btcusd/ohlc"
    }
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: MarketDescr{
				ID:       1,
				Exchange: "bitfinex",
				Pair:     "btcusd",
				Active:   true,
				// NOTE: the routes aren't returned by the API client yet.
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetExchangeMarketsDescr(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/bitfinex"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of exchange's markets index", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetExchangeMarketsDescr("bitfinex")
			},
			resp: `
{
  "result": [
    {
      "id": 1,
      "exchange": "bitfinex",
      "pair": "btcusd",
      "active": true,
      "route": "https://api.cryptowat.ch/markets/bitfinex/btcusd"
    },
    {
      "id": 2,
      "exchange": "bitfinex",
      "pair": "ltcusd",
      "active": true,
      "route": "https://api.cryptowat.ch/markets/bitfinex/ltcusd"
    }
  ],
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: []MarketDescr{
				MarketDescr{
					ID:       1,
					Exchange: "bitfinex",
					Pair:     "btcusd",
					Active:   true,
					Route:    "https://api.cryptowat.ch/markets/bitfinex/btcusd",
				},
				MarketDescr{
					ID:       2,
					Exchange: "bitfinex",
					Pair:     "ltcusd",
					Active:   true,
					Route:    "https://api.cryptowat.ch/markets/bitfinex/ltcusd",
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
