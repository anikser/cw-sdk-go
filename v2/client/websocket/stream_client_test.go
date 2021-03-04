package websocket

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"code.cryptowat.ch/cw-sdk-go/client/websocket/internal"
	"code.cryptowat.ch/cw-sdk-go/common"
	pbm "code.cryptowat.ch/cw-sdk-go/proto/public/markets"
	pbs "code.cryptowat.ch/cw-sdk-go/proto/public/stream"
)

var testStreamSubscriptions = []*StreamSubscription{
	{Resource: "foo"},
	{Resource: "bar"},
}

type receivedError struct {
	err           string
	disconnecting bool
}

func TestStreamClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)

	err := withTestServer(ctx, t, streamServer, func(tp *testServerParams) error {
		client, err := NewStreamClient(&StreamClientParams{
			WSParams: &WSParams{
				URL:       tp.url,
				APIKey:    testApiKey1,
				SecretKey: testSecretKey1,
			},
			Subscriptions: testStreamSubscriptions,
		})
		if err != nil {
			return errors.Trace(err)
		}

		updatesReceived := map[string]bool{
			"OrderBookSnapshot":     false,
			"OrderBookDelta":        false,
			"OrderBookSpreadUpdate": false,
			"TradesUpdate":          false,
			"IntervalsUpdate":       false,
			"SummaryUpdate":         false,
			"SparklineUpdate":       false,
			"VWAPUpdate":            false,
			"VWAPUpdateNano":        false,
			"PerformanceUpdate":     false,
			"TrendlineUpdate":       false,
		}

		updates := make(chan string, len(updatesReceived))

		errorsCh := make(chan receivedError, 32)

		expectError := func(want receivedError) {
			select {
			case got := <-errorsCh:
				assert.Equal(want, got)
			case <-time.After(1 * time.Second):
				t.Errorf("Expected error %q, but got no errors", want.err)
			}
		}

		expectNoErrors := func() {
			select {
			case got := <-errorsCh:
				t.Errorf("Expected no errors, but got %+v", got)
			default:
			}
		}

		// Test market data callbacks

		client.OnMarketUpdate(func(m common.MarketID, md common.MarketUpdate) {
			assert.Equal(m, common.MarketID(1))

			switch {
			case md.OrderBookSnapshot != nil:
				ob := md.OrderBookSnapshot
				assert.Equal(ob.SeqNum, common.SeqNum(testOrderBookUpdate.SeqNum))
				for i, o := range testOrderBookUpdate.Bids {
					bids, err := publicOrderFromProto(o)
					assert.Nil(err)
					assert.Equal(bids, ob.Bids[i])
				}
				for i, o := range testOrderBookUpdate.Asks {
					asks, err := publicOrderFromProto(o)
					assert.Nil(err)
					assert.Equal(asks, ob.Asks[i])
				}
				updates <- "OrderBookSnapshot"

			case md.OrderBookDelta != nil:
				ob := md.OrderBookDelta
				assert.Equal(ob.SeqNum, common.SeqNum(testOrderBookDeltaUpdate.SeqNum))
				for i, o := range testOrderBookDeltaUpdate.Bids.Set {
					set, err := publicOrderFromProto(o)
					assert.Nil(err)
					assert.Equal(set.Price, ob.Bids.Set[i].Price)
					assert.Equal(set.Amount, ob.Bids.Set[i].Amount)
				}
				for i, vs := range testOrderBookDeltaUpdate.Bids.RemoveStr {
					vd, err := decimal.NewFromString(vs)
					assert.Nil(err)
					assert.True(vd.Equal(ob.Bids.Remove[i]))
				}

				for i, o := range testOrderBookDeltaUpdate.Asks.Set {
					set, err := publicOrderFromProto(o)
					assert.Nil(err)
					assert.Equal(set.Price, ob.Asks.Set[i].Price)
					assert.Equal(set.Amount, ob.Asks.Set[i].Amount)
				}
				for i, vs := range testOrderBookDeltaUpdate.Asks.RemoveStr {
					vd, err := decimal.NewFromString(vs)
					assert.Nil(err)
					assert.True(vd.Equal(ob.Asks.Remove[i]))
				}
				updates <- "OrderBookDelta"

			case md.OrderBookSpreadUpdate != nil:
				ob := md.OrderBookSpreadUpdate
				assert.Equal(testOrderBookSpreadUpdate.Timestamp, ob.Timestamp.UnixNano()/int64(time.Millisecond))
				assert.Equal("", cmpStrAndDecimal(testOrderBookSpreadUpdate.Bid.PriceStr, ob.Bid.Price))
				assert.Equal("", cmpStrAndDecimal(testOrderBookSpreadUpdate.Bid.AmountStr, ob.Bid.Amount))
				assert.Equal("", cmpStrAndDecimal(testOrderBookSpreadUpdate.Ask.PriceStr, ob.Ask.Price))
				assert.Equal("", cmpStrAndDecimal(testOrderBookSpreadUpdate.Ask.AmountStr, ob.Ask.Amount))
				updates <- "OrderBookSpreadUpdate"

			case md.TradesUpdate != nil:
				tu := md.TradesUpdate
				for i, t := range testTradesUpdate.Trades {
					if t.Timestamp > 0 {
						assert.Equal(t.Timestamp, tu.Trades[i].Timestamp.Unix())
					}
					if t.TimestampNano > 0 {
						assert.Equal(t.TimestampNano, tu.Trades[i].Timestamp.UnixNano())
					}

					assert.Equal(t.ExternalId, tu.Trades[i].ExternalID)
					assert.Equal("", cmpStrAndDecimal(t.PriceStr, tu.Trades[i].Price))
					assert.Equal("", cmpStrAndDecimal(t.AmountStr, tu.Trades[i].Amount))
				}

				updates <- "TradesUpdate"

			case md.IntervalsUpdate != nil:
				iu := md.IntervalsUpdate
				for i, in := range testIntervalsUpdate.Intervals {
					assert.Equal(in.Closetime, iu.Intervals[i].CloseTime.Unix())
					assert.Equal(in.PeriodName, string(iu.Intervals[i].Period))
					assert.Equal("", cmpStrAndDecimal(in.Ohlc.OpenStr, iu.Intervals[i].OHLC.Open))
					assert.Equal("", cmpStrAndDecimal(in.Ohlc.HighStr, iu.Intervals[i].OHLC.High))
					assert.Equal("", cmpStrAndDecimal(in.Ohlc.LowStr, iu.Intervals[i].OHLC.Low))
					assert.Equal("", cmpStrAndDecimal(in.Ohlc.CloseStr, iu.Intervals[i].OHLC.Close))
					assert.Equal("", cmpStrAndDecimal(in.VolumeBaseStr, iu.Intervals[i].VolumeBase))
					assert.Equal("", cmpStrAndDecimal(in.VolumeQuoteStr, iu.Intervals[i].VolumeQuote))
				}
				updates <- "IntervalsUpdate"

			case md.SummaryUpdate != nil:
				su := md.SummaryUpdate
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.LastStr, su.Last))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.HighStr, su.High))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.LowStr, su.Low))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.VolumeBaseStr, su.VolumeBase))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.VolumeQuoteStr, su.VolumeQuote))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.ChangeAbsoluteStr, su.ChangeAbsolute))
				assert.Equal("", cmpStrAndDecimal(testSummaryUpdate.ChangePercentStr, su.ChangePercent))
				assert.Equal(testSummaryUpdate.NumTrades, su.NumTrades)
				updates <- "SummaryUpdate"

			case md.SparklineUpdate != nil:
				su := md.SparklineUpdate
				assert.Equal(testSparklineUpdate.Time, su.Timestamp.Unix())
				assert.Equal("", cmpStrAndDecimal(testSparklineUpdate.PriceStr, su.Price))
				updates <- "SparklineUpdate"
			}

		})

		var vwapCount int
		client.OnPairUpdate(func(p common.Pair, pd common.PairUpdate) {
			assert.Equal(p.ID, common.PairID("1"))
			switch {
			case pd.VWAPUpdate != nil:
				testVWAPUpdate := testVWAPUpdates[vwapCount]
				fmt.Println(testVWAPUpdate, testVWAPUpdates)
				vu := pd.VWAPUpdate

				switch vwapCount {
				case 0:
					assert.Equal(decimal.NewFromFloat(testVWAPUpdate.Vwap), vu.VWAP)
					assert.Equal(testVWAPUpdate.Timestamp, vu.Timestamp.Unix())
					updates <- "VWAPUpdate"
				case 1:
					assert.Equal(decimal.NewFromFloat(testVWAPUpdate.Vwap), vu.VWAP)
					assert.Equal(testVWAPUpdate.TimestampNano, vu.Timestamp.UnixNano())
					updates <- "VWAPUpdateNano"
				}

				vwapCount++

			case pd.PerformanceUpdate != nil:
				pu := pd.PerformanceUpdate
				assert.Equal(common.PerformanceWindow(testPerformanceUpdate.Window), pu.Window)
				assert.Equal(decimal.NewFromFloat(testPerformanceUpdate.Performance), pu.Performance)
				updates <- "PerformanceUpdate"

			case pd.TrendlineUpdate != nil:
				tu := pd.TrendlineUpdate
				assert.Equal(common.PerformanceWindow(testTrendlineUpdate.Window), tu.Window)
				assert.Equal(decimal.RequireFromString(testTrendlineUpdate.Price), tu.Price)
				assert.Equal(decimal.RequireFromString(testTrendlineUpdate.Volume), tu.Volume)
				updates <- "TrendlineUpdate"
			}
		})

		client.OnError(func(err error, disconnecting bool) {
			v := receivedError{
				err:           err.Error(),
				disconnecting: disconnecting,
			}

			select {
			case errorsCh <- v:
			case <-time.After(1 * time.Second):
				panic(fmt.Sprintf("not able to send to errorsCh: %+v", v))
			}
		})

		if err := client.Connect(); err != nil {
			return errors.Trace(err)
		}

		if err := waitConnOpen(t, tp); err != nil {
			return errors.Errorf("waiting for new conn to be opened: %s", err)
		}

		if err := waitAuthnReq(t, tp, testApiKey1, testSecretKey1); err != nil {
			return errors.Errorf("waiting for authn request: %s", err)
		}

		if err := sendStreamAuthnResp(t, tp, pbs.AuthenticationResult_AUTHENTICATED); err != nil {
			return errors.Errorf("sending authn resp: %s", err)
		}

		subscriptions := []*StreamSubscription{
			// Market subscriptions
			{"markets:1:trades"},
			{"markets:1:summary"},
			{"markets:1:ohlc"},
			{"markets:1:book:snapshots"},
			{"markets:1:book:deltas"},
			{"markets:1:book:spread"},

			// Pair subscriptions
			{"pairs:1:vwap"},
			{"pairs:1:performance"},
			{"pairs:1:trendline"},
		}

		if err := client.Subscribe(subscriptions); err != nil {
			return errors.Errorf("subscribing to baz: %s", err)
		}

		if err := waitSubscribeMsg(t, tp, streamSubsToSubs(subscriptions)); err != nil {
			return errors.Errorf("waiting for subscribe message: %s", err)
		}

		if err := sendStreamUpdates(t, tp); err != nil {
			return errors.Errorf("sending market updates %s", err)
		}

		for i := 0; i < len(updatesReceived); i++ {
			select {
			case update := <-updates:
				if _, alreadyReceived := updatesReceived[update]; !alreadyReceived {
					t.Logf("update processsed %v", update)
					updatesReceived[update] = true
				}

			case <-time.After(1 * time.Second):
				t.Log("Missing updates")
				for u, e := range updatesReceived {
					if !e {
						t.Log(u)
					}
				}
				t.Fail()
			}
		}

		expectNoErrors()

		// Try sending an orderbook update with a malformed price string;
		// expect OnError to be called.
		{
			testOrderBookUpdateWithError := *testOrderBookUpdate
			testOrderBookUpdateWithError.Bids[0].PriceStr = "some_erroneous_price"

			orderBookUpdate := &pbm.MarketUpdateMessage{
				Market: testMarket,
				Update: &pbm.MarketUpdateMessage_OrderBookUpdate{
					OrderBookUpdate: &testOrderBookUpdateWithError,
				},
			}
			if err := sendMarketUpdate(t, tp, orderBookUpdate); err != nil {
				return errors.Trace(err)
			}

			expectError(receivedError{
				err:           "parsing order price: can't convert some_erroneous_price to decimal: exponent is not numeric",
				disconnecting: false,
			})
		}

		return nil
	})
	if err != nil {
		cancel()
		t.Log(errors.ErrorStack(err))
		t.Fatal(err)
	}
}

var testMarket = &pbm.Market{
	ExchangeId:     1,
	CurrencyPairId: 1,
	MarketId:       1,
}

var testOrderBookUpdate = &pbm.OrderBookUpdate{
	SeqNum: 1,
	Bids: []*pbm.Order{
		&pbm.Order{
			PriceStr:  "1.0",
			AmountStr: "100.0",
		},
		&pbm.Order{
			PriceStr:  "2.0",
			AmountStr: "200.0",
		},
	},
	Asks: []*pbm.Order{
		&pbm.Order{
			PriceStr:  "1.5",
			AmountStr: "100.5",
		},
		&pbm.Order{
			PriceStr:  "2.1",
			AmountStr: "200.1",
		},
	},
}

var testOrderBookDeltaUpdate = &pbm.OrderBookDeltaUpdate{
	SeqNum: 1,
	Bids: &pbm.OrderBookDeltaUpdate_OrderDeltas{
		Set: []*pbm.Order{
			&pbm.Order{
				PriceStr:  "1.0",
				AmountStr: "100.0",
			},
		},
		RemoveStr: []string{"30.0"},
	},
	Asks: &pbm.OrderBookDeltaUpdate_OrderDeltas{
		Set: []*pbm.Order{
			&pbm.Order{
				PriceStr:  "4.0",
				AmountStr: "5.0",
			},
		},
		RemoveStr: []string{"31.0"},
	},
}

var testOrderBookSpreadUpdate = &pbm.OrderBookSpreadUpdate{
	Timestamp: 1542223639,
	Bid: &pbm.Order{
		PriceStr:  "1.0",
		AmountStr: "2.0",
	},
	Ask: &pbm.Order{
		PriceStr:  "3.0",
		AmountStr: "4.0",
	},
}

var testTradesUpdate = &pbm.TradesUpdate{
	Trades: []*pbm.Trade{
		&pbm.Trade{
			ExternalId: "1",
			Timestamp:  1542218059,
			PriceStr:   "5.0",
			AmountStr:  "6.0",
		},
		&pbm.Trade{
			ExternalId:    "1",
			TimestampNano: 1542218059000000000,
			PriceStr:      "7.0",
			AmountStr:     "8.0",
		},
		&pbm.Trade{
			ExternalId:    "1",
			TimestampNano: 1542218059000000000,
			PriceStr:      "9.0",
			AmountStr:     "10.0",
		},
	},
}

var testIntervalsUpdate = &pbm.IntervalsUpdate{
	Intervals: []*pbm.Interval{
		&pbm.Interval{
			Closetime:  1542219084,
			PeriodName: "60",
			Ohlc: &pbm.Interval_OHLC{
				OpenStr:  "11.0",
				HighStr:  "12.0",
				LowStr:   "13.0",
				CloseStr: "14.0",
			},
			VolumeBaseStr:  "15.0",
			VolumeQuoteStr: "16.0",
		},
	},
}

var testSummaryUpdate = &pbm.SummaryUpdate{
	LastStr:           "17.0",
	HighStr:           "18.0",
	LowStr:            "19.0",
	VolumeBaseStr:     "20.0",
	VolumeQuoteStr:    "21.0",
	ChangeAbsoluteStr: "22.0",
	ChangePercentStr:  "23.0",
	NumTrades:         1234,
}

var testSparklineUpdate = &pbm.SparklineUpdate{
	Time:     1542205889,
	PriceStr: "25.0",
}

var testPair uint64 = 1

var testVWAPUpdates = []*pbm.PairVwapUpdate{
	{
		Vwap:      1.9,
		Timestamp: 1136239445,
	}, {
		Vwap:          2.1,
		TimestampNano: 1136239445000555999,
	},
}

var testPerformanceUpdate = &pbm.PairPerformanceUpdate{
	Window:      "24h",
	Performance: 1.3,
}

var testTrendlineUpdate = &pbm.PairTrendlineUpdate{
	Window: "24h",
	Time:   1542205889,
	Price:  "1.0",
	Volume: "2.0",
}

func sendStreamUpdates(t *testing.T, tp *testServerParams) error {
	orderBookUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_OrderBookUpdate{
			OrderBookUpdate: testOrderBookUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, orderBookUpdate); err != nil {
		return errors.Trace(err)
	}

	OrderBookDelta := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_OrderBookDeltaUpdate{
			OrderBookDeltaUpdate: testOrderBookDeltaUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, OrderBookDelta); err != nil {
		return errors.Trace(err)
	}

	orderBookSpreadUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_OrderBookSpreadUpdate{
			OrderBookSpreadUpdate: testOrderBookSpreadUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, orderBookSpreadUpdate); err != nil {
		return errors.Trace(err)
	}

	tradesUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_TradesUpdate{
			TradesUpdate: testTradesUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, tradesUpdate); err != nil {
		return errors.Trace(err)
	}

	intervalsUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_IntervalsUpdate{
			IntervalsUpdate: testIntervalsUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, intervalsUpdate); err != nil {
		return errors.Trace(err)
	}

	summaryUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_SummaryUpdate{
			SummaryUpdate: testSummaryUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, summaryUpdate); err != nil {
		return errors.Trace(err)
	}

	sparklineUpdate := &pbm.MarketUpdateMessage{
		Market: testMarket,
		Update: &pbm.MarketUpdateMessage_SparklineUpdate{
			SparklineUpdate: testSparklineUpdate,
		},
	}
	if err := sendMarketUpdate(t, tp, sparklineUpdate); err != nil {
		return errors.Trace(err)
	}

	for _, testVWAPUpdate := range testVWAPUpdates {
		vwapUpdate := &pbm.PairUpdateMessage{
			Pair: testPair,
			Update: &pbm.PairUpdateMessage_VwapUpdate{
				VwapUpdate: testVWAPUpdate,
			},
		}
		if err := sendPairUpdate(t, tp, vwapUpdate); err != nil {
			return errors.Trace(err)
		}
	}

	performanceUpdate := &pbm.PairUpdateMessage{
		Pair: testPair,
		Update: &pbm.PairUpdateMessage_PerformanceUpdate{
			PerformanceUpdate: testPerformanceUpdate,
		},
	}
	if err := sendPairUpdate(t, tp, performanceUpdate); err != nil {
		return errors.Trace(err)
	}

	trendlineUpdate := &pbm.PairUpdateMessage{
		Pair: testPair,
		Update: &pbm.PairUpdateMessage_TrendlineUpdate{
			TrendlineUpdate: testTrendlineUpdate,
		},
	}
	if err := sendPairUpdate(t, tp, trendlineUpdate); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func sendMarketUpdate(t *testing.T, tp *testServerParams, update *pbm.MarketUpdateMessage) error {
	streamMessage := &pbs.StreamMessage{
		Body: &pbs.StreamMessage_MarketUpdate{
			MarketUpdate: update,
		},
	}
	data, err := proto.Marshal(streamMessage)
	if err != nil {
		return errors.Trace(err)
	}

	tp.tx <- internal.WebsocketTx{
		MessageType: websocket.BinaryMessage,
		Data:        data,
	}

	return nil
}

func sendPairUpdate(t *testing.T, tp *testServerParams, update *pbm.PairUpdateMessage) error {
	streamMessage := &pbs.StreamMessage{
		Body: &pbs.StreamMessage_PairUpdate{
			PairUpdate: update,
		},
	}
	data, err := proto.Marshal(streamMessage)
	if err != nil {
		return errors.Trace(err)
	}

	tp.tx <- internal.WebsocketTx{
		MessageType: websocket.BinaryMessage,
		Data:        data,
	}

	return nil
}

// cmpStrAndDecimal tries to interpret str as a decimal, and compares it with
// dec. If parsing str succeeded and numbers are equal, returns an empty
// string; otherwise returns an error message.
func cmpStrAndDecimal(str string, dec decimal.Decimal) string {
	decFromStr, err := decimal.NewFromString(str)
	if err != nil {
		return fmt.Sprintf("failed to parse decimal from %q: %s", str, err)
	}

	if !dec.Equal(decFromStr) {
		return fmt.Sprintf("%q does not equal %q", str, dec.String())
	}

	return ""
}
