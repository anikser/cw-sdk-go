package rest

import (
	"testing"
)

func TestGetPrice(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/kraken/btcusd/price"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of price", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetPrice("kraken", "btcusd")
			},
			resp: `
{
  "result": {
    "price": 9274.34751705
  },
  "allowance": {
    "cost": 1285791,
    "remaining": 7942788866
  }
}`,
			wantResult: dfs("9274.34751705"),
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
