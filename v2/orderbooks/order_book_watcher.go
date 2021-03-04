package orderbooks

import (
	"fmt"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/client/websocket"
	"code.cryptowat.ch/cw-sdk-go/common"
)

// OrderBookWatcherParams are used as options to create a new OrderBookWatcher.
type OrderBookWatcherParams struct {
	Market       common.MarketParams
	RESTClient   *rest.RESTClient
	StreamClient *websocket.StreamClient
}

// OrderBookWatcher uses a RESTClient and StreamClient to keep an order book
// up to date with little setup.
//
// NOTE that it only works best if there is a single OrderBookWatcher per
// StreamClient. Even though it's possible to create multiple
// OrderBookWatcher-s with a single StreamClient, the update handling will be
// very inefficient. If you need to watch multiple markets, it's better to use
// OrderBookUpdater directly, instead of OrderBookWatcher.
type OrderBookWatcher struct {
	market common.Market

	restClient   *rest.RESTClient
	streamClient *websocket.StreamClient
	updater      *OrderBookUpdater
}

// NewOrderBookWatcher creates a new OrderBookWatcher.
func NewOrderBookWatcher(params OrderBookWatcherParams) (*OrderBookWatcher, error) {
	if params.RESTClient == nil {
		panic("RESTClient is nil")
	}

	if params.StreamClient == nil {
		panic("StreamClient is nil")
	}

	market, err := params.RESTClient.GetMarket(params.Market)
	if err != nil {
		return nil, errors.Trace(err)
	}

	snapshotGetter, err := NewOrderBookSnapshotGetterREST(market, params.RESTClient)
	if err != nil {
		return nil, errors.Trace(err)
	}

	updater := NewOrderBookUpdater(&OrderBookUpdaterParams{
		SnapshotGetter: snapshotGetter,
	})

	params.StreamClient.OnMarketUpdate(
		func(updMarketID common.MarketID, md common.MarketUpdate) {
			// If the update is about a different market, ignore it.
			// NOTE that it's very inefficient if a connection has a lot of markets,
			// and it's a lot better for the client to just not use OrderBookWatcher
			// in this case, and use OrderBookUpdater directly instead.
			if updMarketID != market.ID {
				return
			}

			if delta := md.OrderBookDelta; delta != nil {
				updater.ReceiveDelta(*delta)
			}
		},
	)

	params.StreamClient.Subscribe([]*websocket.StreamSubscription{
		&websocket.StreamSubscription{
			Resource: fmt.Sprintf("markets:%d:book:deltas", market.ID),
		},
	})

	ob := &OrderBookWatcher{
		market:       market,
		streamClient: params.StreamClient,
		restClient:   params.RESTClient,
		updater:      updater,
	}

	return ob, nil
}

// Market returns the market for the order book
func (ob *OrderBookWatcher) Market() common.Market {
	return ob.market
}

// GetSnapshot returns a snapshot copy of the current order book.
func (ob *OrderBookWatcher) GetSnapshot() common.OrderBookSnapshot {
	return ob.updater.curOrderBook.GetSnapshot()
}

// OnUpdate calls a callback function whenever the order book updates
func (ob *OrderBookWatcher) OnUpdate(update OnUpdateCB) {
	ob.updater.OnUpdate(update)
}

// Stop stops the order book, along with the websocket connection.
func (ob *OrderBookWatcher) Stop() error {
	ob.streamClient.Unsubscribe([]*websocket.StreamSubscription{
		&websocket.StreamSubscription{
			Resource: fmt.Sprintf("markets:%d:book:deltas", ob.market.ID),
		},
	})

	err := ob.updater.Close()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
