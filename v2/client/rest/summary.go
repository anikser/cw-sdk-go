package rest

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
)

type marketSummariesServer struct {
	Result map[string]marketSummariesItemServer `json:"result"`
}

type marketSummariesItemServer struct {
	Price struct {
		Last   decimal.Decimal `json:"last"`
		High   decimal.Decimal `json:"high"`
		Low    decimal.Decimal `json:"low"`
		Change struct {
			Percentage decimal.Decimal `json:"percentage"`
			Absolute   decimal.Decimal `json:"absolute"`
		} `json:"change"`
	} `json:"price"`
	Volume      decimal.Decimal `json:"volume"`
	VolumeQuote decimal.Decimal `json:"volumeQuote"`
}

type marketSummaryServer struct {
	Result marketSummariesItemServer `json:"result"`
}

func (c *RESTClient) GetSummary(
	exchangeSymbol, pairSymbol string,
) (common.SummaryUpdate, error) {
	result, err := c.do(request{
		endpoint: fmt.Sprintf("markets/%s/%s/summary", exchangeSymbol, pairSymbol),
	})
	if err != nil {
		return common.SummaryUpdate{}, errors.Trace(err)
	}

	srv := marketSummariesItemServer{}

	err = json.Unmarshal(result, &srv)
	if err != nil {
		return common.SummaryUpdate{}, errors.Trace(err)
	}

	ret := marketSummariesItemServerToSummaryUpdate(srv)

	return ret, nil
}

// GetMarketSummaries returns a map from market symbol like "<exchange>:<pair>"
// to the corresponding summary.
func (c *RESTClient) GetMarketSummaries() (map[string]common.SummaryUpdate, error) {
	result, err := c.do(request{
		endpoint: "markets/summaries",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	srv := map[string]marketSummariesItemServer{}

	err = json.Unmarshal(result, &srv)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := make(map[string]common.SummaryUpdate, len(srv))
	for market, v := range srv {
		ret[market] = marketSummariesItemServerToSummaryUpdate(v)
	}

	return ret, nil
}

func marketSummariesItemServerToSummaryUpdate(v marketSummariesItemServer) common.SummaryUpdate {
	return common.SummaryUpdate{
		Last:           v.Price.Last,
		High:           v.Price.High,
		Low:            v.Price.Low,
		ChangeAbsolute: v.Price.Change.Absolute,
		ChangePercent:  v.Price.Change.Percentage,
		VolumeBase:     v.Volume,
		VolumeQuote:    v.VolumeQuote,
	}
}
