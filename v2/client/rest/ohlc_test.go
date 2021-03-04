package rest

import (
	"testing"
	"time"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestGetOHLC(t *testing.T) {
	h := newTestHarnessREST(t, getCheckURL("/markets/kraken/btcusd/ohlc"))

	testCases := []testCaseREST{
		testCaseREST{descr: "Just an example of OHLC data", // {{{
			do: func(c *RESTClient) (interface{}, error) {
				return c.GetOHLC("kraken", "btcusd")
			},
			resp: `
{
  "result": {
    "14400": [
      [
        1559721600,
        6932.7,
        6967.8,
        6838,
        6959.3,
        740.59361951,
        5115702.182713785
      ],
      [
        1559736000,
        6959.3,
        7014.5,
        6887.5,
        6923.3,
        1230.65005308,
        8550003.97889953
      ]
    ],
    "180": [
      [
        1566816660,
        9285.1,
        9293.7,
        9284.7,
        9292.5,
        10.57046452,
        98211.477575253
      ],
      [
        1566816840,
        9292.5,
        9299.9,
        9291.6,
        9295.4,
        3.98015112,
        36995.552742874
      ]
    ]
  },
  "allowance": {
    "cost": 385823,
    "remaining": 7999614177
  }
}`,
			wantResult: map[common.Period][]common.Interval{
				common.Period4H: []common.Interval{
					common.Interval{
						Period:    common.Period4H,
						CloseTime: time.Unix(1559721600, 0).UTC(),
						OHLC: common.OHLC{
							Open:  dfs("6932.7"),
							High:  dfs("6967.8"),
							Low:   dfs("6838"),
							Close: dfs("6959.3"),
						},
						VolumeBase:  dfs("740.59361951"),
						VolumeQuote: dfs("5115702.182713785"),
					},
					common.Interval{
						Period:    common.Period4H,
						CloseTime: time.Unix(1559736000, 0).UTC(),
						OHLC: common.OHLC{
							Open:  dfs("6959.3"),
							High:  dfs("7014.5"),
							Low:   dfs("6887.5"),
							Close: dfs("6923.3"),
						},
						VolumeBase:  dfs("1230.65005308"),
						VolumeQuote: dfs("8550003.97889953"),
					},
				},
				common.Period3M: []common.Interval{
					common.Interval{
						Period:    common.Period3M,
						CloseTime: time.Unix(1566816660, 0).UTC(),
						OHLC: common.OHLC{
							Open:  dfs("9285.1"),
							High:  dfs("9293.7"),
							Low:   dfs("9284.7"),
							Close: dfs("9292.5"),
						},
						VolumeBase:  dfs("10.57046452"),
						VolumeQuote: dfs("98211.477575253"),
					},
					common.Interval{
						Period:    common.Period3M,
						CloseTime: time.Unix(1566816840, 0).UTC(),
						OHLC: common.OHLC{
							Open:  dfs("9292.5"),
							High:  dfs("9299.9"),
							Low:   dfs("9291.6"),
							Close: dfs("9295.4"),
						},
						VolumeBase:  dfs("3.98015112"),
						VolumeQuote: dfs("36995.552742874"),
					},
				},
			},
		},
		// }}}
	}

	h.runTestCases(testCases)
	h.close()
}
