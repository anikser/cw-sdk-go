package rest

import (
	"testing"
)

func TestGetPairsIndex(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/pairs"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of pairs index", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetPairsIndex()
			},
			resp: `
{
  "result": [
    {
      "id": 1365,
      "symbol": "btctusd",
      "base": {
        "id": 60,
        "symbol": "btc",
        "name": "Bitcoin",
        "fiat": false,
        "route": "https://api.cryptowat.ch/assets/btc"
      },
      "quote": {
        "id": 271,
        "symbol": "tusd",
        "name": "True USD",
        "fiat": false,
        "route": "https://api.cryptowat.ch/assets/tusd"
      },
      "route": "https://api.cryptowat.ch/pairs/btctusd"
    },
    {
      "id": 9,
      "symbol": "btcusd",
      "base": {
        "id": 60,
        "symbol": "btc",
        "name": "Bitcoin",
        "fiat": false,
        "route": "https://api.cryptowat.ch/assets/btc"
      },
      "quote": {
        "id": 98,
        "symbol": "usd",
        "name": "United States dollar",
        "fiat": true,
        "route": "https://api.cryptowat.ch/assets/usd"
      },
      "route": "https://api.cryptowat.ch/pairs/btcusd"
    }
  ],
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: []PairDescr{
				PairDescr{
					ID:     1365,
					Symbol: "btctusd",
					Base: PairSideDescr{
						ID:     60,
						Symbol: "btc",
						Name:   "Bitcoin",
						Fiat:   false,
						Route:  "https://api.cryptowat.ch/assets/btc",
					},
					Quote: PairSideDescr{
						ID:     271,
						Symbol: "tusd",
						Name:   "True USD",
						Fiat:   false,
						Route:  "https://api.cryptowat.ch/assets/tusd",
					},
					Route: "https://api.cryptowat.ch/pairs/btctusd",
				},
				PairDescr{
					ID:     9,
					Symbol: "btcusd",
					Base: PairSideDescr{
						ID:     60,
						Symbol: "btc",
						Name:   "Bitcoin",
						Fiat:   false,
						Route:  "https://api.cryptowat.ch/assets/btc",
					},
					Quote: PairSideDescr{
						ID:     98,
						Symbol: "usd",
						Name:   "United States dollar",
						Fiat:   true,
						Route:  "https://api.cryptowat.ch/assets/usd",
					},
					Route: "https://api.cryptowat.ch/pairs/btcusd",
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetPairDescr(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/pairs/btcusd"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of pair description", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetPairDescr("btcusd")
			},
			resp: `
{
  "result": {
    "id": 9,
    "symbol": "btcusd",
    "base": {
      "id": 60,
      "symbol": "btc",
      "name": "Bitcoin",
      "fiat": false,
      "route": "https://api.cryptowat.ch/assets/btc"
    },
    "quote": {
      "id": 98,
      "symbol": "usd",
      "name": "United States dollar",
      "fiat": true,
      "route": "https://api.cryptowat.ch/assets/usd"
    },
    "route": "https://api.cryptowat.ch/pairs/btcusd",
    "markets": [
      {
        "id": 181,
        "exchange": "gemini",
        "pair": "btcusd",
        "active": true,
        "route": "https://api.cryptowat.ch/markets/gemini/btcusd"
      },
      {
        "id": 185,
        "exchange": "quoine",
        "pair": "btcusd",
        "active": true,
        "route": "https://api.cryptowat.ch/markets/quoine/btcusd"
      }
    ]
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: PairDescr{
				ID:     9,
				Symbol: "btcusd",
				Base: PairSideDescr{
					ID:     60,
					Symbol: "btc",
					Name:   "Bitcoin",
					Fiat:   false,
					Route:  "https://api.cryptowat.ch/assets/btc",
				},
				Quote: PairSideDescr{
					ID:     98,
					Symbol: "usd",
					Name:   "United States dollar",
					Fiat:   true,
					Route:  "https://api.cryptowat.ch/assets/usd",
				},
				Route: "https://api.cryptowat.ch/pairs/btcusd",
				Markets: []MarketDescr{
					MarketDescr{
						ID:       181,
						Exchange: "gemini",
						Pair:     "btcusd",
						Active:   true,
						Route:    "https://api.cryptowat.ch/markets/gemini/btcusd",
					},
					MarketDescr{
						ID:       185,
						Exchange: "quoine",
						Pair:     "btcusd",
						Active:   true,
						Route:    "https://api.cryptowat.ch/markets/quoine/btcusd",
					},
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
