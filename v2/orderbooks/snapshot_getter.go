package orderbooks

import (
	"errors"

	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/common"
)

// OrderBookSnapshotGetter gets the up-to-date snapshot. Typically clients
// should use OrderBookSnapshotGetterREST, which gets the snapshot from the
// REST API.
//
// This is needed in the first place because snapshots are broadcasted via
// websocket only every minute, so whenever the client just starts, or gets
// out of sync for a little bit, it needs to get the up-to-date snapshot
// to avoid waiting for it for too long from the websocket.
type OrderBookSnapshotGetter interface {
	GetOrderBookSnapshot() (common.OrderBookSnapshot, error)
}

var _ OrderBookSnapshotGetter = &OrderBookSnapshotGetterREST{}

// OrderBookSnapshotGetterREST implements OrderBookSnapshotGetter; it
// gets snapshot for the specified market from the REST API.
type OrderBookSnapshotGetterREST struct {
	client *rest.RESTClient
	market common.Market
}

// NewOrderBookSnapshotGetterREST creates a new snapshot getter which
// uses the REST API to get snapshots for the given market.
func NewOrderBookSnapshotGetterREST(
	market common.Market,
	restClient *rest.RESTClient,
) (*OrderBookSnapshotGetterREST, error) {
	if restClient == nil {
		return nil, errors.New("RESTClient is nil")
	}
	return &OrderBookSnapshotGetterREST{
		client: restClient,
		market: market,
	}, nil
}

func (sg *OrderBookSnapshotGetterREST) GetOrderBookSnapshot() (common.OrderBookSnapshot, error) {
	return sg.client.GetOrderBookByID(sg.market.ID)
}
