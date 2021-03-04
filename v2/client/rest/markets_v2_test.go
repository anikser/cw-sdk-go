package rest

import (
	"testing"

	"code.cryptowat.ch/cw-sdk-go/common"
)

var bitfinexBTCUSD = common.Market{
	ID: common.MarketID(1),
	Exchange: common.Exchange{
		ID:     common.ExchangeID(1),
		Symbol: "bitfinex",
	},
	Instrument: common.Instrument{
		ID: common.InstrumentID(9),
		Base: common.Asset{
			ID:     60,
			Symbol: "btc",
		},
		Quote: common.Asset{
			ID:     98,
			Symbol: "usd",
		},
	},
}

const bitfinexResponseStr = `
{
  "result": {
    "id": 1,
    "exchange": {
      "id": 1,
      "symbol": "bitfinex"
    },
    "instrument": {
      "id": 9,
      "base": {
        "id": 60,
        "symbol": "btc"
      },
      "quote": {
        "id": 98,
        "symbol": "usd"
      }
    }
  },
  "allowance": {
    "cost": 28898,
    "remaining": 4999971102,
    "upgrade": "Upgrade for a higher allowance, starting at $15/month for 16 seconds/hour. https://cryptowat.ch/pricing"
  }
}`

const bitfinexResponseArrayStr = `
{
  "result": [{
    "id": 1,
    "exchange": {
      "id": 1,
      "symbol": "bitfinex"
    },
    "instrument": {
      "id": 9,
      "base": {
        "id": 60,
        "symbol": "btc"
      },
      "quote": {
        "id": 98,
        "symbol": "usd"
      }
    }
  }],
  "allowance": {
    "cost": 28898,
    "remaining": 4999971102,
    "upgrade": "Upgrade for a higher allowance, starting at $15/month for 16 seconds/hour. https://cryptowat.ch/pricing"
  }
}`

func TestGetMarketByID(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/v2/markets/1"))

	testCases := []testCaseREST{
		testCaseREST{descr: "v2 get market by id",
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetMarketByID(1)
			},
			resp:       bitfinexResponseStr,
			wantResult: bitfinexBTCUSD,
		},
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetMarketBySymbol(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/v2/markets"))

	testCases := []testCaseREST{
		testCaseREST{descr: "v2 get market by symbol",
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetMarketBySymbol(common.MarketSymbol{
					Exchange: "bitfinex",
					Base:     "btc",
					Quote:    "usd",
				})
			},
			resp:       bitfinexResponseArrayStr,
			wantResult: bitfinexBTCUSD,
		},
	}

	h.runTestCases(testCases)
}
