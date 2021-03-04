package rest

import (
	"testing"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestGetExchangeBySymbol(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/v2/exchanges"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of exchanges index", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetExchangeBySymbol(common.ExchangeSymbol("kraken"))
			},
			resp: `
{
  "result": {
    "id": 4,
    "symbol": "kraken"
  },
  "allowance": {
    "cost": 16717,
    "remaining": 3999983283,
    "remainingPaid": 0
  }
}`,
			wantResult: common.Exchange{
				ID:     4,
				Symbol: "kraken",
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetExchangesIndex(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/exchanges"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of exchanges index", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetExchangesIndex()
			},
			resp: `
{
  "result": [
    {
      "id": 63,
      "symbol": "hitbtc",
      "name": "HitBTC",
      "route": "https://api.cryptowat.ch/exchanges/hitbtc",
      "active": true
    },
    {
      "id": 2,
      "symbol": "coinbase-pro",
      "name": "Coinbase Pro",
      "route": "https://api.cryptowat.ch/exchanges/coinbase-pro",
      "active": true
    }
  ],
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: []ExchangeDescrBrief{
				ExchangeDescrBrief{
					ID:     63,
					Symbol: "hitbtc",
					Name:   "HitBTC",
					Route:  "https://api.cryptowat.ch/exchanges/hitbtc",
					Active: true,
				},
				ExchangeDescrBrief{
					ID:     2,
					Symbol: "coinbase-pro",
					Name:   "Coinbase Pro",
					Route:  "https://api.cryptowat.ch/exchanges/coinbase-pro",
					Active: true,
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetExchangeDescr(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/exchanges/kraken"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of exchange description", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetExchangeDescr("kraken")
			},
			resp: `
{
  "result": {
    "id": 4,
    "symbol": "kraken",
    "name": "Kraken",
    "active": true,
    "routes": {
      "markets": "https://api.cryptowat.ch/markets/kraken"
    }
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: &ExchangeDescr{
				ID:     4,
				Symbol: "kraken",
				Name:   "Kraken",
				Routes: ExchangeDescrRoutes{
					Markets: "https://api.cryptowat.ch/markets/kraken",
				},
				Active: true,
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
