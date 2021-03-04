package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestCache(t *testing.T) {
	c := New()

	var (
		marketID    = common.MarketID(11)
		xchSymbol   = common.ExchangeSymbol("kraken")
		baseSymbol  = common.AssetSymbol("btc")
		quoteSymbol = common.AssetSymbol("usd")
		assetID     = common.AssetID(22)
	)

	a1 := common.Asset{
		ID:     assetID,
		Symbol: baseSymbol,
	}

	m1 := common.Market{
		ID: marketID,
		Exchange: common.Exchange{
			Symbol: xchSymbol,
		},
		Instrument: common.Instrument{
			Base: common.Asset{
				Symbol: baseSymbol,
			},
			Quote: common.Asset{
				Symbol: quoteSymbol,
			},
		},
	}

	_, hit := c.GetMarketByID(marketID)
	assert.Equal(t, false, hit)

	_, hit = c.GetAssetByID(assetID)
	assert.Equal(t, false, hit)

	c.SetMarket(m1)
	c.SetAsset(a1)

	mc, hit := c.GetMarketByID(marketID)
	assert.Equal(t, true, hit)
	assert.Equal(t, m1, mc)

	mc, hit = c.GetMarketBySymbol(m1.Symbol())
	assert.Equal(t, true, hit)
	assert.Equal(t, m1, mc)

	ac, hit := c.GetAssetByID(assetID)
	assert.Equal(t, true, hit)
	assert.Equal(t, a1, ac)
}
