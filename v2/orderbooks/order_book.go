package orderbooks

import (
	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/common"
)

// OrderBook represents a mutable order book, which is able to receive snapshots
// and deltas.
//
// It is not thread-safe; so if you need to use it from more than one
// goroutine, apply your own synchronization.
type OrderBook struct {
	snapshot common.OrderBookSnapshot
}

func NewOrderBook(snapshot common.OrderBookSnapshot) *OrderBook {
	return &OrderBook{
		snapshot: snapshot,
	}
}

// GetSnapshot returns the snapshot of the current orderbook.
func (ob *OrderBook) GetSnapshot() common.OrderBookSnapshot {
	return ob.snapshot
}

// GetSeqNum is a shortcut for GetSnapshot().SeqNum
func (ob *OrderBook) GetSeqNum() common.SeqNum {
	return ob.snapshot.SeqNum
}

// ApplyDelta applies the given delta (received from the wire) to the current
// orderbook. If the sequence number isn't exactly the old one incremented by
// 1, returns an error without applying delta.
func (ob *OrderBook) ApplyDelta(obd common.OrderBookDelta) error {
	return ob.ApplyDeltaOpt(obd, false)
}

// ApplyDeltaOpt applies the given delta (received from the wire) to the
// current orderbook. If ignoreSeqNum is true, applies the delta even if the
// sequence number isn't exactly the old one incremented by 1.
func (ob *OrderBook) ApplyDeltaOpt(obd common.OrderBookDelta, ignoreSeqNum bool) error {
	newSnapshot, err := ob.snapshot.ApplyDeltaOpt(obd, ignoreSeqNum)
	if err != nil {
		return errors.Trace(err)
	}

	ob.snapshot = newSnapshot

	return nil
}

// ApplySnapshot sets the internal orderbook to the provided snapshot.
func (ob *OrderBook) ApplySnapshot(snapshot common.OrderBookSnapshot) {
	ob.snapshot = snapshot
}
