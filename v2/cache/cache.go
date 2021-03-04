package cache

import (
	"code.cryptowat.ch/cw-sdk-go/common"
)

type Cache struct {
	marketsBySymbol map[common.MarketSymbol]common.Market
	marketsByID     map[common.MarketID]common.Market
	assetsBySymbol  map[common.AssetSymbol]common.Asset
	assetsByID      map[common.AssetID]common.Asset
}

func (c *Cache) GetMarketBySymbol(ms common.MarketSymbol) (
	common.Market, bool,
) {
	m, hit := c.marketsBySymbol[ms]
	return m, hit
}

func (c *Cache) GetMarketByID(id common.MarketID) (common.Market, bool) {
	m, hit := c.marketsByID[id]
	return m, hit
}

func (c *Cache) GetAssetBySymbol(symbol common.AssetSymbol) (common.Asset, bool) {
	a, hit := c.assetsBySymbol[symbol]
	return a, hit
}

func (c *Cache) GetAssetByID(id common.AssetID) (common.Asset, bool) {
	m, hit := c.assetsByID[id]
	return m, hit
}

type marketSymbol struct {
	exchange common.ExchangeSymbol
	base     common.AssetSymbol
	quote    common.AssetSymbol
}

func (c *Cache) SetMarket(m common.Market) {
	c.marketsBySymbol[m.Symbol()] = m
	c.marketsByID[m.ID] = m
	c.SetAsset(m.Instrument.Base)
	c.SetAsset(m.Instrument.Quote)
}

func (c *Cache) SetAsset(a common.Asset) {
	c.assetsBySymbol[a.Symbol] = a
	c.assetsByID[a.ID] = a
}

func New() *Cache {
	return &Cache{
		marketsBySymbol: make(map[common.MarketSymbol]common.Market),
		marketsByID:     make(map[common.MarketID]common.Market),
		assetsBySymbol:  make(map[common.AssetSymbol]common.Asset),
		assetsByID:      make(map[common.AssetID]common.Asset),
	}
}
