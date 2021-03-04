// Package cw allows sdk users to easily get Cryptowatch objects without doing any extra work.
package cw

import "code.cryptowat.ch/cw-sdk-go/common"

// GetMarketParams is the parameter type to GetMarket which allows you to get a market by
// Symbol OR ID. Only one needs to be present.
type GetMarketParams struct {
	Symbol common.MarketSymbol
	ID     common.MarketID
}

// GetAssetParams is the parameter type to GetAsset which allows you to get an asset by
// Symbol or ID. Only one needs to be present.
type GetAssetParams struct {
	Symbol common.AssetSymbol
	ID     common.AssetID
}
