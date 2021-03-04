package common

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// SeqNum is used to make sure order book deltas are processed in order. Each order book delta update has
// the sequence number incremented by 1. Sometimes sequence numbers reset to a smaller value (this happens when we deploy or
// restart our back end for whatever reason). In those cases, the first broadcasted update is a snapshot, so clients should
// not assume they are out of sync when they receive a snapshot with a lower seq number.
type SeqNum uint32

// OrderBookSnapshot represents a full order book snapshot.
type OrderBookSnapshot struct {
	// SeqNum	is the sequence number of the last order book delta received.
	// Since snapshots are broadcast on a 1-minute interval regardless of updates,
	// it's possible this value doesn't change.
	// See the SeqNum definition for more information.
	SeqNum SeqNum

	Bids []PublicOrder
	Asks []PublicOrder
}

func (obs *OrderBookSnapshot) Copy() OrderBookSnapshot {
	bids := make([]PublicOrder, len(obs.Bids))
	asks := make([]PublicOrder, len(obs.Asks))
	copy(bids, obs.Bids)
	copy(asks, obs.Asks)

	return OrderBookSnapshot{
		SeqNum: obs.SeqNum,
		Bids:   bids,
		Asks:   asks,
	}
}

func (s *OrderBookSnapshot) Empty() bool {
	return len(s.Bids) == 0 && len(s.Asks) == 0
}

func (s *OrderBookSnapshot) IsValid() bool {
	if len(s.Bids) == 0 || len(s.Asks) == 0 {
		return true
	}

	if s.Bids[0].Price.GreaterThanOrEqual(s.Asks[0].Price) {
		return false
	}

	return true
}

func (s *OrderBookSnapshot) ApplyDelta(obd OrderBookDelta) (OrderBookSnapshot, error) {
	return s.ApplyDeltaOpt(obd, false)
}

func (s *OrderBookSnapshot) ApplyDeltaOpt(obd OrderBookDelta, ignoreSeqNum bool) (OrderBookSnapshot, error) {
	// Refuse to apply delta of there is a gap in sequence numbers
	if !ignoreSeqNum && obd.SeqNum-1 != s.SeqNum {
		return OrderBookSnapshot{}, ErrSeqNumMismatch
	}

	return OrderBookSnapshot{
		Bids:   ApplyDeltas(s.Bids, &obd.Bids, true),
		Asks:   ApplyDeltas(s.Asks, &obd.Asks, false),
		SeqNum: obd.SeqNum,
	}, nil
}

// OrderBookDelta represents an order book delta update, which is
// the minimum amount of data necessary to keep a local order book up to date.
// Since order book snapshots are throttled at 1 per minute, subscribing to
// the delta updates is the best way to keep an order book up to date.
type OrderBookDelta struct {
	// SeqNum is used to make sure deltas are processed in order.
	// See the SeqNum definition for more information.
	SeqNum SeqNum

	Bids OrderDeltas
	Asks OrderDeltas
}

// Empty returns whether OrderBookDelta doesn't contain any deltas for bids and
// asks.
func (delta OrderBookDelta) Empty() bool {
	return delta.Bids.Empty() && delta.Asks.Empty()
}

// OrderDeltas are used to update an order book, either by setting (adding)
// new PublicOrders, or removing orders at specific prices.
type OrderDeltas struct {
	// Set is a list of orders used to add or replace orders on an order book.
	// Each order in Set is guaranteed to be a different price. For each of them,
	// if the order at that price exists on the book, replace it. If an order
	// at that price does not exist, add it to the book.
	Set []PublicOrder

	// Remove is a list of prices. To apply to an order book, remove all orders
	// of that price from that book.
	Remove []decimal.Decimal
}

// Empty returns whether the OrderDeltas doesn't contain any deltas.
func (d OrderDeltas) Empty() bool {
	return len(d.Set) == 0 && len(d.Remove) == 0
}

// OrderBookSpreadUpdate represents the most recent order book spread. It
// consists of the best current bid and as price.
type OrderBookSpreadUpdate struct {
	Timestamp time.Time
	Bid       PublicOrder
	Ask       PublicOrder
}

// Applying orderbook deltas {{{

type orderDeltaAction int

const (
	orderDeltaActionSet orderDeltaAction = iota
	orderDeltaActionRemove
)

type orderDelta struct {
	action orderDeltaAction
	order  PublicOrder

	index      int
	overridden bool
}

// ApplyDeltas applies given deltas to the slice of orders, and returns a newly
// allocated slice of orders, sorted by price accordingly to reverse argument.
func ApplyDeltas(orders []PublicOrder, deltas *OrderDeltas, reverse bool) []PublicOrder {
	// Flatten deltas, i.e. make all sets and removes in a single sortable list {{{
	ditems := make([]orderDelta, 0, len(deltas.Set)+len(deltas.Remove))
	for i, v := range deltas.Set {
		ditems = append(ditems, orderDelta{
			action: orderDeltaActionSet,
			order:  v,
			index:  i,
		})
	}

	for i, v := range deltas.Remove {
		ditems = append(ditems, orderDelta{
			action: orderDeltaActionRemove,
			order: PublicOrder{
				Price: v,
			},
			index: len(deltas.Set) + i,
		})
	}
	// }}}

	// Sort them by price (accordingly to reverse flag) and by index {{{
	// Also set the overridden flag for all items with the same price, except
	// the one with largest index.
	sort.Slice(ditems, func(i, j int) bool {
		cmp := ditems[i].order.Price.Cmp(ditems[j].order.Price)

		if cmp != 0 {
			// Return whether it's "less" accordingly to the reverse flag
			return reverse != (cmp < 0)
		}

		if ditems[i].index < ditems[j].index {
			ditems[i].overridden = true
			return true
		} else {
			ditems[j].overridden = true
			return false
		}
	})
	// }}}

	// Right now we have a list of deltas, sorted by price, and therefore by
	// index in the snapshot at which they should be applied.

	// Allocate a new slice of orders. Set capacity so that it's certainly
	// enough to avoid reallocs, even though might be more than required.
	newOrders := make([]PublicOrder, 0, len(orders)+len(ditems))

	curOrdersIdx := 0

	// Apply deltas {{{
	for _, delta := range ditems {
		if delta.overridden {
			continue
		}

		idx, exists := binarySearch(orders, delta.order.Price, curOrdersIdx, reverse)

		// Copy all orders up to the found idx, they're unchanged
		newOrders = append(newOrders, orders[curOrdersIdx:idx]...)

		// Apply current order delta
		switch delta.action {
		case orderDeltaActionSet:
			newOrders = append(newOrders, delta.order)
			if exists {
				// Replace (as opposed to Insert), so we also need to increment
				// the old snapshot's index
				idx += 1
			}
		case orderDeltaActionRemove:
			if !exists {
				// Need to remove an order, but it doesn't exist, so it's a no-op.
				// TODO: set some flag that delta wasn't applied cleanly
				break
			}

			// Just skip an order
			idx += 1
		}

		curOrdersIdx = idx
	}
	// }}}

	// Copy the remainder of old orders, they are unchanged
	newOrders = append(newOrders, orders[curOrdersIdx:]...)

	return newOrders
}

// binarySearch performs binary search in a sorted slice of orders for the
// given price, in the range [start, len(orders)). If the order is reversed,
// the reversed flag must be set to true.
func binarySearch(orders []PublicOrder, price decimal.Decimal, start int, reversed bool) (idx int, exists bool) {
	end := len(orders)

	for start < end {
		curIdx := start + (end-start)/2
		order := orders[curIdx]
		cmp := price.Cmp(order.Price)
		if cmp == 0 {
			// Found the item
			return curIdx, true
		} else {
			// Didn't find the item yet, divide the range.

			if reversed != (cmp < 0) {
				// It's "less" accordingly to the reverse flag
				end = curIdx
			} else {
				start = curIdx + 1
			}
		}
	}

	// Item doesn't exist, and start is where it can be inserted
	return start, false
}

// }}}

// Generating orderbook deltas {{{

func GetDeltasAgainst(oldOrders, newOrders []PublicOrder, reversed bool) OrderDeltas {
	oldIdx := 0
	newIdx := 0

	delta := OrderDeltas{}

	for oldIdx < len(oldOrders) && newIdx < len(newOrders) {
		oldOrder := oldOrders[oldIdx]
		newOrder := newOrders[newIdx]

		cmp := oldOrder.Price.Cmp(newOrder.Price)
		if cmp == 0 {
			// Prices are equal, let's see if amounts are different
			if !oldOrder.Amount.Equal(newOrder.Amount) {
				// Amounts are different, so the Set delta is needed
				delta.Set = append(delta.Set, PublicOrder{
					Price:  newOrder.Price,
					Amount: newOrder.Amount,
				})
			}

			oldIdx += 1
			newIdx += 1
		} else if reversed != (cmp < 0) {
			// Old order is smaller than the new one: need to add Remove delta
			delta.Remove = append(delta.Remove, oldOrder.Price)

			oldIdx += 1
		} else {
			// New order is smaller than the old one: need to add Set delta
			delta.Set = append(delta.Set, PublicOrder{
				Price:  newOrder.Price,
				Amount: newOrder.Amount,
			})

			newIdx += 1
		}
	}

	// If there are some old orders remaining, add Remove delta for each of them
	for _, oldOrder := range oldOrders[oldIdx:] {
		delta.Remove = append(delta.Remove, oldOrder.Price)
	}

	// If there are some new orders remaining, add Set delta for each of them
	for _, newOrder := range newOrders[newIdx:] {
		delta.Set = append(delta.Set, PublicOrder{
			Price:  newOrder.Price,
			Amount: newOrder.Amount,
		})
	}

	return delta
}

// GetDeltasAgainst creates deltas which would have to be applied to
// oldSnapshot to get s.
func (s *OrderBookSnapshot) GetDeltasAgainst(oldSnapshot OrderBookSnapshot) OrderBookDelta {
	return OrderBookDelta{
		Bids:   GetDeltasAgainst(oldSnapshot.Bids, s.Bids, true),
		Asks:   GetDeltasAgainst(oldSnapshot.Asks, s.Asks, false),
		SeqNum: s.SeqNum,
	}
}

// }}}
