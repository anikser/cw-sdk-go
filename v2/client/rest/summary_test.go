package rest

import (
	"testing"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestGetSummary(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/kraken/btcusd/summary"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of summary data", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetSummary("kraken", "btcusd")
			},
			resp: `
{
  "result": {
    "price": {
      "last": 9175.7,
      "high": 9365,
      "low": 9039.4,
      "change": {
        "percentage": -0.0125373969565872,
        "absolute": -116.5
      }
    },
    "volume": 2971.81319108,
    "volumeQuote": 27255819.6957213
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: common.SummaryUpdate{
				Last:           dfs("9175.7"),
				High:           dfs("9365"),
				Low:            dfs("9039.4"),
				ChangePercent:  dfs("-0.0125373969565872"),
				ChangeAbsolute: dfs("-116.5"),
				VolumeBase:     dfs("2971.81319108"),
				VolumeQuote:    dfs("27255819.6957213"),
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}

func TestGetMarketSummaries(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/summaries"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of all summary data", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetMarketSummaries()
			},
			resp: `
{
  "result": {
    "kraken:btcusd": {
      "price": {
        "last": 9175.7,
        "high": 9365,
        "low": 9039.4,
        "change": {
          "percentage": -0.0125373969565872,
          "absolute": -116.5
        }
      },
      "volume": 2971.81319108,
      "volumeQuote": 27255819.6957213
    },
    "binance:adaeth": {
      "price": {
        "last": 0.00026184,
        "high": 0.00027006,
        "low": 0.00025888,
        "change": {
          "percentage": -0.0080315199272617,
          "absolute": -2.12e-06
        }
      },
      "volume": 4186959,
      "volumeQuote": 1100.83486215
    }
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: map[string]common.SummaryUpdate{
				"kraken:btcusd": common.SummaryUpdate{
					Last:           dfs("9175.7"),
					High:           dfs("9365"),
					Low:            dfs("9039.4"),
					ChangePercent:  dfs("-0.0125373969565872"),
					ChangeAbsolute: dfs("-116.5"),
					VolumeBase:     dfs("2971.81319108"),
					VolumeQuote:    dfs("27255819.6957213"),
				},
				"binance:adaeth": common.SummaryUpdate{
					Last:           dfs("0.00026184"),
					High:           dfs("0.00027006"),
					Low:            dfs("0.00025888"),
					ChangePercent:  dfs("-0.0080315199272617"),
					ChangeAbsolute: dfs("-2.12e-06"),
					VolumeBase:     dfs("4186959"),
					VolumeQuote:    dfs("1100.83486215"),
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
