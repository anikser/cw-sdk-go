package websocket

import (
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
	pbb "code.cryptowat.ch/cw-sdk-go/proto/public/broker"
	pbc "code.cryptowat.ch/cw-sdk-go/proto/public/client"
	pbm "code.cryptowat.ch/cw-sdk-go/proto/public/markets"
	pbs "code.cryptowat.ch/cw-sdk-go/proto/public/stream"
)

func publicOrderFromProto(po *pbm.Order) (common.PublicOrder, error) {
	price, err := decimal.NewFromString(po.PriceStr)
	if err != nil {
		return common.PublicOrder{}, errors.Annotatef(err, "parsing order price")
	}

	amount, err := decimal.NewFromString(po.AmountStr)
	if err != nil {
		return common.PublicOrder{}, errors.Annotatef(err, "parsing order amount")
	}

	return common.PublicOrder{
		Price:  price,
		Amount: amount,
	}, nil
}

func orderBookSnapshotUpdateFromProto(obu *pbm.OrderBookUpdate) (common.OrderBookSnapshot, error) {
	bids := make([]common.PublicOrder, len(obu.Bids))
	for i, b := range obu.Bids {
		var err error
		bids[i], err = publicOrderFromProto(b)
		if err != nil {
			return common.OrderBookSnapshot{}, errors.Trace(err)
		}
	}

	asks := make([]common.PublicOrder, len(obu.Asks))
	for i, a := range obu.Asks {
		var err error
		asks[i], err = publicOrderFromProto(a)
		if err != nil {
			return common.OrderBookSnapshot{}, errors.Trace(err)
		}
	}

	return common.OrderBookSnapshot{
		SeqNum: common.SeqNum(obu.SeqNum),
		Bids:   bids,
		Asks:   asks,
	}, nil
}

func orderBookDeltasFromProto(deltas *pbm.OrderBookDeltaUpdate_OrderDeltas) (common.OrderDeltas, error) {
	set := make([]common.PublicOrder, len(deltas.Set))
	for i, o := range deltas.Set {
		var err error
		set[i], err = publicOrderFromProto(o)
		if err != nil {
			return common.OrderDeltas{}, errors.Trace(err)
		}
	}

	remove := make([]decimal.Decimal, len(deltas.RemoveStr))
	for i, v := range deltas.RemoveStr {
		var err error
		remove[i], err = decimal.NewFromString(v)
		if err != nil {
			return common.OrderDeltas{}, errors.Annotatef(err, "parsing remove delta price")
		}
	}

	return common.OrderDeltas{
		Set:    set,
		Remove: remove,
	}, nil
}

func orderBookDeltaUpdateFromProto(obdu *pbm.OrderBookDeltaUpdate) (common.OrderBookDelta, error) {
	deltaBids, err := orderBookDeltasFromProto(obdu.Bids)
	if err != nil {
		return common.OrderBookDelta{}, errors.Trace(err)
	}

	deltaAsks, err := orderBookDeltasFromProto(obdu.Asks)
	if err != nil {
		return common.OrderBookDelta{}, errors.Trace(err)
	}

	return common.OrderBookDelta{
		SeqNum: common.SeqNum(obdu.SeqNum),

		Bids: deltaBids,
		Asks: deltaAsks,
	}, nil
}

func orderBookSpreadUpdateFromProto(obsu *pbm.OrderBookSpreadUpdate) (common.OrderBookSpreadUpdate, error) {
	bid, err := publicOrderFromProto(obsu.Bid)
	if err != nil {
		return common.OrderBookSpreadUpdate{}, errors.Trace(err)
	}

	ask, err := publicOrderFromProto(obsu.Ask)
	if err != nil {
		return common.OrderBookSpreadUpdate{}, errors.Trace(err)
	}

	return common.OrderBookSpreadUpdate{
		Timestamp: time.Unix(0, obsu.Timestamp*int64(time.Millisecond)),
		Bid:       bid,
		Ask:       ask,
	}, nil
}

func tradesUpdateFromProto(tu *pbm.TradesUpdate) (common.TradesUpdate, error) {
	pt := make([]common.PublicTrade, len(tu.Trades))

	for i, t := range tu.Trades {
		var timestamp time.Time

		// Eliminate timestamp redundancy
		if t.TimestampNano > 0 {
			timestamp = time.Unix(0, t.TimestampNano)
		} else {
			timestamp = time.Unix(t.Timestamp, 0)
		}

		price, err := decimal.NewFromString(t.PriceStr)
		if err != nil {
			return common.TradesUpdate{}, errors.Annotatef(err, "parsing delta price")
		}

		amount, err := decimal.NewFromString(t.AmountStr)
		if err != nil {
			return common.TradesUpdate{}, errors.Annotatef(err, "parsing delta amount")
		}

		pt[i] = common.PublicTrade{
			ExternalID: t.ExternalId,
			Timestamp:  timestamp,
			Price:      price,
			Amount:     amount,
			OrderSide:  orderSideFromPublicTradesProto(t.OrderSide),
		}
	}

	return common.TradesUpdate{
		Trades: pt,
	}, nil
}

func orderSideFromPublicTradesProto(ts pbm.Trade_OrderSide) common.OrderSide {
	if ts == pbm.Trade_BUYSIDE {
		return common.OrderSideBuy
	}
	if ts == pbm.Trade_SELLSIDE {
		return common.OrderSideSell
	}
	return common.OrderSideUnknown
}

func intervalsUpdateFromProto(iu *pbm.IntervalsUpdate) (common.IntervalsUpdate, error) {
	var is []common.Interval

	for _, i := range iu.Intervals {
		v, err := intervalFromProto(i)
		if err != nil {
			return common.IntervalsUpdate{}, errors.Trace(err)
		}
		is = append(is, v)
	}

	return common.IntervalsUpdate{
		Intervals: is,
	}, nil
}

func intervalFromProto(i *pbm.Interval) (common.Interval, error) {
	open, err := decimal.NewFromString(i.Ohlc.OpenStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC open")
	}

	high, err := decimal.NewFromString(i.Ohlc.HighStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC high")
	}

	low, err := decimal.NewFromString(i.Ohlc.LowStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC low")
	}

	close, err := decimal.NewFromString(i.Ohlc.CloseStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC close")
	}

	volumeBase, err := decimal.NewFromString(i.VolumeBaseStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC volumeBase")
	}

	volumeQuote, err := decimal.NewFromString(i.VolumeQuoteStr)
	if err != nil {
		return common.Interval{}, errors.Annotatef(err, "parsing OHLC volumeQuote")
	}

	return common.Interval{
		CloseTime: time.Unix(i.Closetime, 0),
		Period:    common.Period(i.PeriodName),
		OHLC: common.OHLC{
			Open:  open,
			High:  high,
			Low:   low,
			Close: close,
		},
		VolumeBase:  volumeBase,
		VolumeQuote: volumeQuote,
	}, nil
}

func summaryUpdateFromProto(su *pbm.SummaryUpdate) (common.SummaryUpdate, error) {
	last, err := decimal.NewFromString(su.LastStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary last price")
	}

	high, err := decimal.NewFromString(su.HighStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary high")
	}

	low, err := decimal.NewFromString(su.LowStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary low")
	}

	volumeBase, err := decimal.NewFromString(su.VolumeBaseStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary volumeBase")
	}

	volumeQuote, err := decimal.NewFromString(su.VolumeQuoteStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary volumeQuote")
	}

	changeAbsolute, err := decimal.NewFromString(su.ChangeAbsoluteStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary changeAbsolute")
	}

	changePercent, err := decimal.NewFromString(su.ChangePercentStr)
	if err != nil {
		return common.SummaryUpdate{}, errors.Annotatef(err, "parsing summary changePercent")
	}

	return common.SummaryUpdate{
		Last:           last,
		High:           high,
		Low:            low,
		VolumeBase:     volumeBase,
		VolumeQuote:    volumeQuote,
		ChangeAbsolute: changeAbsolute,
		ChangePercent:  changePercent,
		NumTrades:      su.NumTrades,
	}, nil
}

func sparklineUpdateFromProto(su *pbm.SparklineUpdate) (common.SparklineUpdate, error) {
	price, err := decimal.NewFromString(su.PriceStr)
	if err != nil {
		return common.SparklineUpdate{}, errors.Annotatef(err, "parsing sparkline price")
	}

	return common.SparklineUpdate{
		Timestamp: time.Unix(su.Time, 0),
		Price:     price,
	}, nil
}

func vwapUpdateFromProto(vu *pbm.PairVwapUpdate) common.VWAPUpdate {
	// Eliminate timestamp redundancy
	var t time.Time
	if vu.TimestampNano > 0 {
		t = time.Unix(0, vu.TimestampNano)
	} else {
		t = time.Unix(vu.Timestamp, 0)
	}

	return common.VWAPUpdate{
		VWAP:      decimal.NewFromFloat(vu.Vwap),
		Timestamp: t,
	}
}

func performanceUpdateFromProto(pu *pbm.PairPerformanceUpdate) common.PerformanceUpdate {
	return common.PerformanceUpdate{
		Window:      common.PerformanceWindow(pu.Window),
		Performance: decimal.NewFromFloat(pu.Performance),
	}
}

func trendlineUpdateFromProto(tu *pbm.PairTrendlineUpdate) (common.TrendlineUpdate, error) {
	price, err := decimal.NewFromString(tu.Price)
	if err != nil {
		return common.TrendlineUpdate{}, errors.Annotatef(err, "parsing trendline price")
	}

	volume, err := decimal.NewFromString(tu.Volume)
	if err != nil {
		return common.TrendlineUpdate{}, errors.Annotatef(err, "parsing trendline volume")
	}

	return common.TrendlineUpdate{
		Window:    common.PerformanceWindow(tu.Window),
		Timestamp: time.Unix(tu.Time, 0),
		Price:     price,
		Volume:    volume,
	}, nil
}

func privateOrderFromProto(order *pbb.PrivateOrder) common.PrivateOrder {
	var priceParams common.PriceParams
	for _, pp := range order.PriceParams {
		priceParams = append(priceParams, &common.PriceParam{
			Value: pp.ValueString,
			Type:  common.PriceParamType(pp.Type),
		})
	}

	if order.AmountFilledString == "" {
		order.AmountFilledString = "0.0"
	}

	return common.PrivateOrder{
		Timestamp:   time.Unix(order.Time, 0),
		PriceParams: priceParams,
		Amount:      order.AmountParamString,
		OrderSide:   common.OrderSide(order.Side),
		OrderType:   common.OrderType(order.Type),
		FundingType: common.FundingType(order.FundingType),
		ExpireTime:  time.Unix(order.ExpireTime, 0),

		ID:           order.Id,
		Leverage:     order.Leverage,
		CurrentStop:  order.CurrentStopString,
		InitialStop:  order.InitialStopString,
		AmountFilled: order.AmountFilledString,
	}
}

func placeOrderParamsToProto(orderOpt common.PlaceOrderParams) *pbb.PrivateOrder {
	priceParams := make([]*pbb.PrivateOrder_PriceParam, 0, len(orderOpt.PriceParams))
	for _, p := range orderOpt.PriceParams {
		priceParams = append(priceParams, &pbb.PrivateOrder_PriceParam{
			ValueString: p.Value,
			Type:        pbb.PrivateOrder_PriceParamType(p.Type),
		})
	}

	return &pbb.PrivateOrder{
		Side:              int32(orderOpt.OrderSide),
		Type:              pbb.PrivateOrder_Type(orderOpt.OrderType),
		FundingType:       pbb.FundingType(orderOpt.FundingType),
		PriceParams:       priceParams,
		AmountParamString: orderOpt.Amount,
		Leverage:          orderOpt.Leverage,
		ExpireTime:        timeToUnix(orderOpt.ExpireTime),
	}
}

func privateOrderToProto(o common.PrivateOrder) *pbb.PrivateOrder {
	priceParams := []*pbb.PrivateOrder_PriceParam{}
	for _, p := range o.PriceParams {
		priceParams = append(priceParams, &pbb.PrivateOrder_PriceParam{
			ValueString: p.Value,
			Type:        pbb.PrivateOrder_PriceParamType(p.Type),
		})
	}

	return &pbb.PrivateOrder{
		Id:          o.ID,
		Side:        int32(o.OrderSide),
		Type:        pbb.PrivateOrder_Type(o.OrderType),
		FundingType: pbb.FundingType(o.FundingType),
		Time:        o.Timestamp.Unix(),
		PriceParams: priceParams,

		AmountParamString:  o.Amount,
		AmountFilledString: o.AmountFilled,
		ExpireTime:         timeToUnix(o.ExpireTime),
		Leverage:           o.Leverage,
	}
}

func balancesToProto(balances common.Balances) []*pbb.Balances {
	balancesByFType := map[pbb.FundingType][]*pbb.Balance{}
	for _, xchBalances := range balances {
		for _, balance := range xchBalances {
			fType := pbb.FundingType(balance.FundingType)
			balancesByFType[fType] = append(balancesByFType[fType], &pbb.Balance{
				Currency:     balance.Asset.String(),
				AmountString: balance.Amount.String(),
			})
		}
	}

	ret := []*pbb.Balances{}
	for fundingType, balances := range balancesByFType {
		ret = append(ret, &pbb.Balances{
			FundingType: fundingType,
			Balances:    balances,
		})
	}

	return ret
}

func positionFromProto(position *pbb.PrivatePosition) common.PrivatePosition {
	return common.PrivatePosition{
		ExternalID:   position.Id,
		Timestamp:    time.Unix(position.Time, 0),
		OrderSide:    common.OrderSide(position.Side),
		AvgPrice:     position.AvgPriceString,
		AmountOpen:   position.AmountOpenString,
		AmountClosed: position.AmountClosedString,
		OrderIDs:     position.OrderIds,
		TradeIDs:     position.TradeIds,
	}
}

func positionToProto(p common.PrivatePosition) *pbb.PrivatePosition {
	return &pbb.PrivatePosition{
		Id:                 p.ExternalID,
		Time:               p.Timestamp.Unix(),
		Side:               int32(p.OrderSide),
		AvgPriceString:     p.AvgPrice,
		AmountOpenString:   p.AmountOpen,
		AmountClosedString: p.AmountClosed,
		OrderIds:           p.OrderIDs,
		TradeIds:           p.TradeIDs,
	}
}

func tradeFromProto(trade *pbb.PrivateTrade) common.PrivateTrade {
	// Eliminate timestamp redundancy
	var t time.Time
	t = time.Unix(trade.Time, 0)

	return common.PrivateTrade{
		ExternalID: trade.ExternalId,
		OrderID:    trade.OrderId,
		Timestamp:  t,
		OrderSide:  common.OrderSide(trade.Side),
		Price:      trade.PriceString,
		Amount:     trade.AmountString,
	}
}

func tradeToProto(t common.PrivateTrade) *pbb.PrivateTrade {
	return &pbb.PrivateTrade{
		ExternalId:   t.ExternalID,
		OrderId:      t.OrderID,
		Time:         t.Timestamp.Unix(),
		PriceString:  t.Price,
		AmountString: t.Amount,
		Side:         int32(t.OrderSide),
	}
}

func subscriptionResultFromProto(sr *pbs.SubscriptionResult) SubscriptionResult {
	failed := make([]SubscribeError, 0, len(sr.Failed))
	for _, v := range sr.Failed {
		failed = append(failed, SubscribeError{
			Key:          v.Key,
			Error:        v.Error,
			Subscription: subFromProto(v.GetSubscription()),
		})
	}

	return SubscriptionResult{
		Subscriptions: subsFromProto(sr.GetSubscriptions()),
		Failed:        failed,
		Status: SubscriptionStatus{
			Subscriptions: subsFromProto(sr.Status.GetSubscriptions()),
		},
	}
}

func unsubscriptionResultFromProto(sr *pbs.UnsubscriptionResult) UnsubscriptionResult {
	failed := make([]UnsubscribeError, 0, len(sr.Failed))
	for _, v := range sr.Failed {
		failed = append(failed, UnsubscribeError{
			Key:          v.Key,
			Error:        v.Error,
			Subscription: subFromProto(v.GetSubscription()),
		})
	}

	return UnsubscriptionResult{
		Subscriptions: subsFromProto(sr.GetSubscriptions()),
		Failed:        failed,
		Status: SubscriptionStatus{
			Subscriptions: subsFromProto(sr.Status.GetSubscriptions()),
		},
	}
}

func bandwidthFromProto(bu *pbs.BandwidthUpdate) Bandwidth {
	return Bandwidth{
		OK:             bu.Ok,
		BytesRemaining: bu.BytesRemaining,
		BytesUsed:      bu.BytesUsed,
	}
}

func missedMessagesFromProto(mm *pbs.MissedMessages) MissedMessages {
	return MissedMessages{
		NumMissedMessages: mm.NumMissedMessages,
	}
}

func unmarshalAuthnResultStream(data []byte) (*pbs.AuthenticationResult, error) {
	var msg pbs.StreamMessage

	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, errors.Trace(err)
	}

	authnResult := msg.GetAuthenticationResult()
	if authnResult == nil {
		return nil, errors.Trace(ErrInvalidAuthn)
	}

	return authnResult, nil
}

func unmarshalAuthnResultTrade(data []byte) (*pbs.AuthenticationResult, error) {
	var msg pbb.BrokerUpdateMessage

	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, errors.Trace(err)
	}

	authnResult := msg.GetAuthenticationResult()
	if authnResult == nil {
		return nil, errors.Trace(ErrInvalidAuthn)
	}

	return authnResult, nil
}

func subsToProto(subs []Subscription) []*pbc.ClientSubscription {
	if len(subs) == 0 {
		return nil
	}

	switch subs[0].(type) {
	case *StreamSubscription:
		return streamSubsToProto(subs)
	case *tradeSubscription:
		return tradeSubsToProto(subs)
	}

	panic(errInvalidSubType)
}

func subsFromProto(subs []*pbc.ClientSubscription) []Subscription {
	if len(subs) == 0 {
		return nil
	}

	switch subs[0].Body.(type) {
	case *pbc.ClientSubscription_StreamSubscription:
		return streamSubsToSubs(streamSubsFromProto(subs))
	case *pbc.ClientSubscription_TradeSubscription:
		return tradeSubsToSubs(tradeSubsFromProto(subs))
	}

	panic(errInvalidSubType)
}

func subFromProto(sub *pbc.ClientSubscription) Subscription {
	if sub == nil {
		return &StreamSubscription{}
	}
	switch v := sub.Body.(type) {
	case *pbc.ClientSubscription_StreamSubscription:
		return streamSubFromProto(v)
	case *pbc.ClientSubscription_TradeSubscription:
		return tradeSubFromProto(v)
	}

	panic(errInvalidSubType)
}

func streamSubToProto(sub *StreamSubscription) *pbc.ClientSubscription_StreamSubscription {
	return &pbc.ClientSubscription_StreamSubscription{
		StreamSubscription: &pbc.StreamSubscription{
			Resource: sub.Resource,
		},
	}
}

func streamSubFromProto(sub *pbc.ClientSubscription_StreamSubscription) *StreamSubscription {
	if sub == nil {
		return &StreamSubscription{}
	}

	return &StreamSubscription{
		Resource: sub.StreamSubscription.Resource,
	}
}

func tradeSubFromProto(sub *pbc.ClientSubscription_TradeSubscription) *tradeSubscription {
	if sub == nil {
		return &tradeSubscription{}
	}

	var auth *ExchangeAuth
	if ta := sub.TradeSubscription.GetAuth(); ta != nil {
		auth = &ExchangeAuth{
			APIKey:        ta.ApiKey,
			APISecret:     ta.ApiSecret,
			CustomerID:    ta.CustomerId,
			KeyPassphrase: ta.KeyPassphrase,
		}
	}

	marketIDStr := sub.TradeSubscription.GetMarketId()
	marketID, err := strconv.Atoi(marketIDStr)
	if err != nil {
		// This should never happen, it indicates there is a problem with the backend
		panic("invalid market id")
	}

	return &tradeSubscription{
		marketID: common.MarketID(marketID),
		auth:     auth,
	}
}

func streamSubsToProto(subs []Subscription) []*pbc.ClientSubscription {
	ret := make([]*pbc.ClientSubscription, 0, len(subs))

	for _, sub := range subs {
		v, ok := sub.(*StreamSubscription)
		if !ok {
			panic(errInvalidSubType)
		}

		ret = append(ret, &pbc.ClientSubscription{
			Body: streamSubToProto(v),
		})
	}

	return ret
}

func streamSubsFromProto(subs []*pbc.ClientSubscription) []*StreamSubscription {
	ret := make([]*StreamSubscription, 0, len(subs))

	for _, s := range subs {
		v, ok := s.Body.(*pbc.ClientSubscription_StreamSubscription)
		if !ok {
			panic(errInvalidSubType)
		}

		ret = append(ret, streamSubFromProto(v))
	}

	return ret
}

func tradeSubsToProto(subs []Subscription) []*pbc.ClientSubscription {
	ret := make([]*pbc.ClientSubscription, 0, len(subs))

	for _, sub := range subs {
		v, ok := sub.(*tradeSubscription)
		if !ok {
			panic(errInvalidSubType)
		}

		var auth *pbc.TradeSessionAuth

		if v.auth != nil {
			auth = &pbc.TradeSessionAuth{
				ApiKey:        v.auth.APIKey,
				ApiSecret:     v.auth.APISecret,
				CustomerId:    v.auth.CustomerID,
				KeyPassphrase: v.auth.KeyPassphrase,
			}
		}

		ret = append(ret, &pbc.ClientSubscription{
			Body: &pbc.ClientSubscription_TradeSubscription{
				TradeSubscription: &pbc.TradeSubscription{
					MarketId: v.marketID.String(),
					Auth:     auth,
				},
			},
		})
	}

	return ret
}

func tradeSubsFromProto(subs []*pbc.ClientSubscription) []*tradeSubscription {
	ret := make([]*tradeSubscription, 0, len(subs))

	for _, s := range subs {
		v, ok := s.Body.(*pbc.ClientSubscription_TradeSubscription)
		if !ok {
			panic(errInvalidSubType)
		}

		ret = append(ret, tradeSubFromProto(v))
	}

	return ret
}

func unixMillisToTime(unixMillis int64) time.Time {
	secs := unixMillis / 1000
	ns := (unixMillis - secs*1000 /*ms*/) * 1000 /*ns*/
	return time.Unix(secs, ns)
}

// timeToUnix is like t.Unix(), but it returns 0 if t has zero value.
func timeToUnix(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}

func float32ToString(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}

func float64ToString(v float64) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 64)
}

func uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}
