package rest

import (
	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/cache"
	"code.cryptowat.ch/cw-sdk-go/common"
)

type MockV2Client struct {
	C *cache.Cache
}

func NewMockV2Client() *MockV2Client {
	return &MockV2Client{
		C: cache.New(),
	}
}

func (mc *MockV2Client) GetMarketBySymbol(ms common.MarketSymbol) (
	common.Market, error,
) {
	m, _ := mc.C.GetMarketBySymbol(ms)
	return m, nil
}

func (mc *MockV2Client) GetMarketByID(id common.MarketID) (
	common.Market, error,
) {
	m, _ := mc.C.GetMarketByID(id)
	return m, nil
}

func (mc *MockV2Client) GetAssetBySymbol(symbol common.AssetSymbol) (
	common.Asset, error,
) {
	a, _ := mc.C.GetAssetBySymbol(symbol)
	return a, nil
}

func (mc *MockV2Client) GetAssetByID(id common.AssetID) (
	common.Asset, error,
) {
	a, _ := mc.C.GetAssetByID(id)
	return a, nil
}

func (mc *MockV2Client) GetOrderBookByID(id common.MarketID) (
	common.OrderBookSnapshot, error,
) {
	return common.OrderBookSnapshot{}, errors.New("Not implemented")
}
