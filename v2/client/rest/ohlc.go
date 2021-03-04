package rest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
)

type ohlcServer struct {
	Result map[string][][]decimal.Decimal `json:"result"`
}

// GetOHLC returns a map from a period (represented as a number of seconds: 60,
// 180, etc) to a slice of OHLC candles for that period, in the time ascending
// order.
func (c *RESTClient) GetOHLC(
	exchangeSymbol, pairSymbol string,
) (map[common.Period][]common.Interval, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s/ohlc", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	srv := map[string][][]decimal.Decimal{}
	err = json.Unmarshal(result, &srv)

	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := make(map[common.Period][]common.Interval, len(srv))
	for srvPeriod, srvCandles := range srv {
		period := common.Period(srvPeriod)
		candles := make([]common.Interval, 0, len(srvCandles))
		for _, srvCandle := range srvCandles {
			if len(srvCandle) < 7 {
				return nil, errors.Errorf("unexpected response from the server: wanted 7 elements, got %v", srvCandle)
			}

			ts := srvCandle[0].IntPart()

			candle := common.Interval{
				Period:    period,
				CloseTime: time.Unix(ts, 0).UTC(),
				OHLC: common.OHLC{
					Open:  srvCandle[1],
					High:  srvCandle[2],
					Low:   srvCandle[3],
					Close: srvCandle[4],
				},
				VolumeBase:  srvCandle[5],
				VolumeQuote: srvCandle[6],
			}

			candles = append(candles, candle)
		}

		ret[period] = candles
	}

	return ret, nil
}
