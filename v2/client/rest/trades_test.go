package rest

import (
	"testing"
	"time"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestGetTrades(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/kraken/btcusd/trades"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of trades", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetTrades("kraken", "btcusd")
			},
			resp: `
{
  "result": [
    [
      0,
      1568110135,
      9313.4,
      0.23510492
    ],
    [
      0,
      1568110136,
      9314.1,
      0.005
    ],
    [
      0,
      1568110150,
      9312.3,
      0.020304
    ]
  ],
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: []common.PublicTrade{
				common.PublicTrade{
					Timestamp: time.Unix(1568110135, 0),
					Price:     dfs("9313.4"),
					Amount:    dfs("0.23510492"),
				},
				common.PublicTrade{
					Timestamp: time.Unix(1568110136, 0),
					Price:     dfs("9314.1"),
					Amount:    dfs("0.005"),
				},
				common.PublicTrade{
					Timestamp: time.Unix(1568110150, 0),
					Price:     dfs("9312.3"),
					Amount:    dfs("0.020304"),
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
