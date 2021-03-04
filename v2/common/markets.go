package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"
)

var (
	// ErrSeqNumMismatch is returned from OrderBookSnapshot.ApplyDelta... family
	// when new squence number isn't exactly the old one plus 1.
	ErrSeqNumMismatch = errors.New("seq num mismatch")
)

type MarketID int

func (m MarketID) String() string {
	return strconv.Itoa(int(m))
}

type MarketSymbol struct {
	Exchange ExchangeSymbol
	Base     AssetSymbol
	Quote    AssetSymbol
}

func (ms MarketSymbol) String() string {
	return fmt.Sprintf("%s:%s%s", ms.Exchange, ms.Base, ms.Quote)
}

type ExchangeID int
type ExchangeSymbol string

type InstrumentID int

type AssetID int

func (a AssetID) String() string {
	return strconv.Itoa(int(a))
}

type AssetSymbol string

func (as AssetSymbol) String() string {
	return string(as)
}

// Market represents a market on Cryptowatch. A market consists of an exchange
// and currency pair. For example, Kraken BTC/USD. IDs are used instead of
// names to avoid inconsistencies associated with name changes.
// Example: kraken:btcusd, bitfinex:btcusd
type Market struct {
	ID         MarketID   `json:"id"`
	Exchange   Exchange   `json:"exchange"`
	Instrument Instrument `json:"instrument"`
}

func (m Market) Symbol() MarketSymbol {
	return MarketSymbol{
		Exchange: m.Exchange.Symbol,
		Base:     m.Instrument.Base.Symbol,
		Quote:    m.Instrument.Quote.Symbol,
	}
}

func (m Market) String() string {
	return m.Symbol().String()
}

type MarketParams struct {
	Symbol MarketSymbol
	ID     MarketID
}

// Exchange represents an exchange on Cryptowatch.
// Examples: kraken, bitfinex, coinbase-pro
type Exchange struct {
	ID     ExchangeID     `json:"id"`
	Symbol ExchangeSymbol `json:"symbol"`
}

// Instrument represents an instrument on Cryptowatch, which is essentially a base+quote,
// or if it is a futures market, a settlement frequency as well.
// Examples: btcusd, ethusd, btcusdt
type Instrument struct {
	ID    InstrumentID `json:"id"`
	Base  Asset        `json:"base"`
	Quote Asset        `json:"quote"`
}

type GetAssetParams struct {
	ID     AssetID
	Symbol AssetSymbol
}

// An asset represents an asset on Cryptowatch. Assets are the building blocks for instruments.
// Examples: btc, usd, eth
type Asset struct {
	ID     AssetID     `json:"id"`
	Symbol AssetSymbol `json:"symbol"`
}

func (a Asset) String() string {
	return string(a.Symbol)
}

// MarketUpdate is a container for all market data callbacks. For any MarketUpdate
// intance, it will only ever have one of its properties non-null.
// See OnMarketUpdate.
type MarketUpdate struct {
	OrderBookSnapshot     *OrderBookSnapshot     `json:"OrderBookSnapshot,omitempty"`
	OrderBookDelta        *OrderBookDelta        `json:"OrderBookDelta,omitempty"`
	OrderBookSpreadUpdate *OrderBookSpreadUpdate `json:"OrderBookSpreadUpdate,omitempty"`
	TradesUpdate          *TradesUpdate          `json:"TradesUpdate,omitempty"`
	IntervalsUpdate       *IntervalsUpdate       `json:"IntervalsUpdate,omitempty"`
	SummaryUpdate         *SummaryUpdate         `json:"SummaryUpdate,omitempty"`
	SparklineUpdate       *SparklineUpdate       `json:"SparklineUpdate,omitempty"`
}

func (v MarketUpdate) String() string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("[failed to stringify MarketUpdate: %s]", err)
	}

	return string(data)
}

// PublicOrder represents a public order placed on an exchange. They often come
// as a slice of PublicOrder, which come with order book updates.
type PublicOrder struct {
	Price  decimal.Decimal
	Amount decimal.Decimal
}

// TradesUpdate represents the most recent trades that have occurred for a
// particular market.
type TradesUpdate struct {
	Trades []PublicTrade
}

// PublicTrade represents a trade made on an exchange. See TradesUpdate and
// OnTradesUpdate.
type PublicTrade struct {
	// ExternalID is given by the exchange, and not Cryptowatch. NOTE some of
	// these may be "0" since we still use int64 on the back end to represent
	// these IDs, and some are strings.
	ExternalID string
	Timestamp  time.Time
	Price      decimal.Decimal
	Amount     decimal.Decimal
	OrderSide  OrderSide
}

// An interval represetns OHLC data for a particular Period and CloseTime.
type Interval struct {
	Period Period
	OHLC   OHLC

	// CloseTime is the time at which the Interval ended.
	CloseTime time.Time

	// VolumeBase is the amount of volume traded over this Interval, represented
	// in the base currency.
	VolumeBase decimal.Decimal

	// VolumeQuote is the amount of volume traded over this Interval, represented
	// in the quote currency.
	VolumeQuote decimal.Decimal
}

// Period is the number of seconds in an Interval.
type Period string

// The following constants are all available Period values for IntervalsUpdate.
const (
	Period1M         Period = "60"
	Period3M         Period = "180"
	Period5M         Period = "300"
	Period15M        Period = "900"
	Period30M        Period = "1800"
	Period1H         Period = "3600"
	Period2H         Period = "7200"
	Period4H         Period = "14400"
	Period6H         Period = "21600"
	Period12H        Period = "43200"
	Period1D         Period = "86400"
	Period3D         Period = "259200"
	Period1WThursday Period = "604800"
	Period1WMonday   Period = "604800_Monday"
)

func (p Period) Duration() time.Duration {
	return periodDurations[p]
}

var periodDurations = map[Period]time.Duration{
	"60":            60 * time.Second,
	"180":           180 * time.Second,
	"300":           300 * time.Second,
	"900":           900 * time.Second,
	"1800":          1800 * time.Second,
	"3600":          3600 * time.Second,
	"7200":          7200 * time.Second,
	"14400":         14400 * time.Second,
	"21600":         21600 * time.Second,
	"43200":         43200 * time.Second,
	"86400":         86400 * time.Second,
	"259200":        259200 * time.Second,
	"604800":        604800 * time.Second,
	"604800_Monday": 604800 * time.Second,
}

// PeriodNames contains human-readable names for Period.
// e.g. Period1M = 60 (seconds) = "1m".
var PeriodNames = map[Period]string{
	Period1M:         "1m",
	Period3M:         "3m",
	Period5M:         "5m",
	Period15M:        "15m",
	Period30M:        "30m",
	Period1H:         "1h",
	Period2H:         "2h",
	Period4H:         "4h",
	Period6H:         "6h",
	Period12H:        "12h",
	Period1D:         "1d",
	Period3D:         "3d",
	Period1WMonday:   "1w",
	Period1WThursday: "1w",
}

// OHLC contains the open, low, high, and close prices for a given Interval.
type OHLC struct {
	Open  decimal.Decimal
	High  decimal.Decimal
	Low   decimal.Decimal
	Close decimal.Decimal
}

// IntervalsUpdate represents an update for OHLC data, at all relevant periods.
// Intervals is represented as a slice because more than one interval update
// can occur at the same time for different periods. For example, Period1M and
// Period3M will overlap every 3 minutes, so that IntervalsUpdate will contain
// both intervals.
type IntervalsUpdate struct {
	Intervals []Interval
}

// SummaryUpdate represents recent summary information for a particular market.
type SummaryUpdate struct {
	Last           decimal.Decimal
	High           decimal.Decimal
	Low            decimal.Decimal
	VolumeBase     decimal.Decimal
	VolumeQuote    decimal.Decimal
	ChangeAbsolute decimal.Decimal
	ChangePercent  decimal.Decimal
	NumTrades      int32
}

// SparklineUpdate represents the sparkline update for a market at a particular time.
// A sparkline is a very small line chart, typically drawn without axes or coordinates.
// It presents the general shape of the variation in price.
// https://en.wikipedia.org/wiki/Sparkline
type SparklineUpdate struct {
	Timestamp time.Time
	Price     decimal.Decimal
}
