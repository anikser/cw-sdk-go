package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// OrderSide represents the order side; e.g. "buy" or "sell".
type OrderSide int32

func (os OrderSide) String() string {
	return OrderSideNames[os]
}

const (
	OrderSideUnknown OrderSide = iota
	OrderSideBuy
	OrderSideSell
)

// OrderSideNames contains human-readable names for OrderSide.
var OrderSideNames = map[OrderSide]string{
	OrderSideSell:    "sell",
	OrderSideBuy:     "buy",
	OrderSideUnknown: "unknown",
}

// OrderType represents the type of order; e.g. "market" or "limit". There are
// 13 different types of orders. Those available depend on the exchange. Refer
// to the exchange's documentation for what order types are available.
type OrderType int32

// The following constants define all possible order types.
const (
	MarketOrder OrderType = iota
	LimitOrder
	StopLossOrder
	StopLossLimitOrder
	TakeProfitOrder
	TakeProfitLimitOrder
	StopLossTakeProfitOrder
	StopLossTakeProfitLimitOrder
	TrailingStopLossOrder
	TrailingStopLossLimitOrder
	StopLossAndLimitOrder
	FillOrKillOrder
	SettlePositionOrder
)

// OrderTypeNames contains human-readable names for OrderType.
var OrderTypeNames = map[OrderType]string{
	MarketOrder:                  "market",
	LimitOrder:                   "limit",
	StopLossOrder:                "stop-loss",
	TakeProfitOrder:              "take-profit",
	TakeProfitLimitOrder:         "take-profit-limit",
	StopLossTakeProfitOrder:      "stop-loss-take-profit",
	StopLossTakeProfitLimitOrder: "stop-loss-take-profit-limit",
	TrailingStopLossOrder:        "trailing-stop-loss",
	TrailingStopLossLimitOrder:   "trailing-stop-loss-limit",
	StopLossAndLimitOrder:        "stop-loss-and-limit",
	FillOrKillOrder:              "fill-or-kill",
	SettlePositionOrder:          "settle-position",
}

// FundingType represents the funding type for an order; e.g. "spot" or "margin".
// The funding types available depend on the exchange, as well as your account's
// permissions on the exchange.
type FundingType int32

// The following constants define every possible funding type.
const (
	SpotFunding FundingType = iota
	MarginFunding
	FuturesFunding
)

// FundingTypeNames contains human-readable names for FundingType.
var FundingTypeNames = map[FundingType]string{
	SpotFunding:    "spot",
	MarginFunding:  "margin",
	FuturesFunding: "futures",
}

// PriceParam is used as input for an Order.
type PriceParam struct {
	Value string
	Type  PriceParamType
}

// PriceParamType represents the type of price parameter used in PriceParams.
type PriceParamType int32

// The following constants define every possible PriceParamType
const (
	AbsoluteValuePrice PriceParamType = iota
	RelativeValuePrice
	RelativePercentValuePrice
)

// PriceParams is a list of price parameters that define the input to an order.
// Usually you will just need one PriceParam input, but some order types take
// multiple PriceParam inputs, such as TrailingStopLossOrder.
// TODO document different PriceParam uses
type PriceParams []*PriceParam

// PlaceOrderParams contains the necessary options for creating a new order with
// the trade client.
// See TradeClient.PlaceOrder.
type PlaceOrderParams struct {
	MarketID    MarketID
	PriceParams PriceParams
	Amount      string
	OrderSide   OrderSide
	OrderType   OrderType
	FundingType FundingType
	Leverage    string
	ExpireTime  time.Time
}

// CancelOrderParams contains the necessary options for canceling an existing order with
// the trade client.
// See TradeClient.CancelOrder.
type CancelOrderParams struct {
	MarketID MarketID
	OrderID  string
}

// PrivateOrder represents an order you have placed on an exchange, either
// through the TradeClient, or on the exchange itself.
type PrivateOrder struct {
	PriceParams PriceParams
	Amount      string
	OrderSide   OrderSide
	OrderType   OrderType
	FundingType FundingType
	ExpireTime  time.Time

	// Set by server and updated internally by client.
	// ID previously was ExternalID.
	ID           string
	Timestamp    time.Time
	Leverage     string
	CurrentStop  string
	InitialStop  string
	AmountFilled string

	// Broker error code; 0 if successful
	Error int32
}

// String implmements the fmt.Stringer interface for PrivateOrder.
func (o PrivateOrder) String() string {
	var priceStr string
	if o.OrderType == MarketOrder {
		priceStr = "n/a"
	} else {
		for _, p := range o.PriceParams {
			priceStr += p.Value + " "
		}
	}

	var expiresStr string
	if o.ExpireTime.IsZero() {
		expiresStr = "n/a"
	} else {
		expiresStr = fmt.Sprintf("%v", o.ExpireTime)
	}

	return fmt.Sprintf("[%v] [%s - %s/%s] id=%s amount=%s amount_filled=%v value=%s expires=%v",
		o.Timestamp, OrderSideNames[o.OrderSide], FundingTypeNames[o.FundingType],
		OrderTypeNames[o.OrderType], o.ID, o.Amount, o.AmountFilled, priceStr, expiresStr,
	)
}

// CacheKey returns the key composed by joining the market ID with the ID.
func (o PrivateOrder) CacheKey(mID MarketID) string {
	return strings.Join([]string{string(mID), o.ID}, "_")
}

// PrivateTrade represents a trade made on your account.
type PrivateTrade struct {
	ExternalID string
	OrderID    string
	Timestamp  time.Time
	Price      string
	Amount     string
	OrderSide  OrderSide
}

// PrivatePosition represents one of your positions on an exchange.
type PrivatePosition struct {
	ExternalID   string
	Timestamp    time.Time
	OrderSide    OrderSide
	AvgPrice     string
	AmountOpen   string
	AmountClosed string
	OrderIDs     []string
	TradeIDs     []string
}

// Balance is the amount you have of a particular asset.
type Balance struct {
	FundingType FundingType
	Asset       Asset
	Amount      decimal.Decimal
}

type Balances map[Exchange][]Balance

// All returns a slice of all balances combined based on funding type + asset
func (b Balances) All() []Balance {
	ret := []Balance{}

	type cacheIndex struct {
		asset Asset
		ftype FundingType
	}

	cache := map[cacheIndex]Balance{}
	for _, balances := range b {
		for _, balance := range balances {
			ci := cacheIndex{
				asset: balance.Asset,
				ftype: balance.FundingType,
			}
			if _, ok := cache[ci]; ok {
				cache[ci].Amount.Add(balance.Amount)
				continue
			}
			cache[ci] = Balance{
				FundingType: balance.FundingType,
				Asset:       balance.Asset,
				Amount:      balance.Amount,
			}
		}
	}

	for _, bal := range cache {
		ret = append(ret, bal)
	}

	return ret
}

type PrivateOrders []PrivateOrder

func (os PrivateOrders) Len() int {
	return len(os)
}

func (os PrivateOrders) Less(i, j int) bool {
	return os[j].Timestamp.After(os[i].Timestamp)
}

func (os PrivateOrders) Swap(i, j int) {
	os[i], os[j] = os[j], os[i]
}

type PrivateTrades []PrivateTrade

func (ts PrivateTrades) Len() int {
	return len(ts)
}

func (ts PrivateTrades) Less(i, j int) bool {
	return ts[j].Timestamp.After(ts[i].Timestamp)
}

func (ts PrivateTrades) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

type PrivatePositions []PrivatePosition

func (ps PrivatePositions) Len() int {
	return len(ps)
}

func (ps PrivatePositions) Less(i, j int) bool {
	return ps[j].Timestamp.After(ps[i].Timestamp)
}

func (ps PrivatePositions) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
