package rest

import (
	"encoding/json"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/common"
)

type V2 interface {
	GetMarketBySymbol(common.MarketSymbol) (common.Market, error)
	GetMarketByID(common.MarketID) (common.Market, error)
	GetAssetByID(common.AssetID) (common.Asset, error)
	GetAssetBySymbol(common.AssetSymbol) (common.Asset, error)
	GetOrderBookByID(common.MarketID) (common.OrderBookSnapshot, error)
}

// GetMarket returns the Market based on an ID or symbol
func (c *RESTClient) GetMarket(params common.MarketParams) (
	common.Market, error,
) {
	if params.ID > 0 {
		return c.GetMarketByID(params.ID)
	}
	return c.GetMarketBySymbol(params.Symbol)
}

// GetMarketBySymbol returns Market object based on exchange, base, and quote.
func (c *RESTClient) GetMarketBySymbol(params common.MarketSymbol) (
	common.Market, error,
) {
	result, err := c.do(request{
		endpoint: "v2/markets",
		params: map[string]string{
			"exchange": string(params.Exchange),
			"base":     string(params.Base),
			"quote":    string(params.Quote),
		},
	})
	if err != nil {
		return common.Market{}, errors.Trace(err)
	}

	marketResp := [1]common.Market{}
	err = json.Unmarshal(result, &marketResp)

	return marketResp[0], errors.Trace(err)
}

// GetMarketByID returns a Market object based on the market's ID.
func (c *RESTClient) GetMarketByID(id common.MarketID) (
	common.Market, error,
) {
	result, err := c.do(request{
		endpoint: "v2/markets/" + id.String(),
	})
	if err != nil {
		return common.Market{}, errors.Trace(err)
	}

	market := common.Market{}
	err = json.Unmarshal(result, &market)

	return market, errors.Trace(err)
}

func (c *RESTClient) GetAssetByID(id common.AssetID) (
	common.Asset, error,
) {
	result, err := c.do(request{
		endpoint: "v2/assets/" + id.String(),
	})
	if err != nil {
		return common.Asset{}, errors.Trace(err)
	}

	asset := common.Asset{}
	err = json.Unmarshal(result, &asset)

	return asset, errors.Trace(err)
}

func (c *RESTClient) GetAssetBySymbol(symbol common.AssetSymbol) (
	common.Asset, error,
) {
	result, err := c.do(request{
		endpoint: "v2/assets",
		params: map[string]string{
			"symbol": string(symbol),
		},
	})
	if err != nil {
		return common.Asset{}, errors.Trace(err)
	}

	asset := common.Asset{}
	err = json.Unmarshal(result, &asset)

	return asset, errors.Trace(err)
}
