package orderbooks

import (
	"fmt"
	"testing"

	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/common"
)

func TestOrderBook(t *testing.T) {
	if err := testOrderBook(t); err != nil {
		t.Fatal(errors.ErrorStack(err))
	}
}

func testOrderBook(t *testing.T) error {
	var err error
	ob := NewOrderBook(common.OrderBookSnapshot{
		SeqNum: 1,
		Bids: []common.PublicOrder{
			common.PublicOrder{Price: dfs("100"), Amount: dfs("1")},
			common.PublicOrder{Price: dfs("92"), Amount: dfs("2")},
			common.PublicOrder{Price: dfs("91"), Amount: dfs("2")},
			common.PublicOrder{Price: dfs("90"), Amount: dfs("3")},
		},
		Asks: []common.PublicOrder{
			common.PublicOrder{Price: dfs("110"), Amount: dfs("4")},
			common.PublicOrder{Price: dfs("120"), Amount: dfs("5")},
		},
	})

	err = compareSnapshots(
		common.OrderBookSnapshot{
			SeqNum: 1,
			Bids: []common.PublicOrder{
				common.PublicOrder{Price: dfs("100"), Amount: dfs("1")},
				common.PublicOrder{Price: dfs("92"), Amount: dfs("2")},
				common.PublicOrder{Price: dfs("91"), Amount: dfs("2")},
				common.PublicOrder{Price: dfs("90"), Amount: dfs("3")},
			},
			Asks: []common.PublicOrder{
				common.PublicOrder{Price: dfs("110"), Amount: dfs("4")},
				common.PublicOrder{Price: dfs("120"), Amount: dfs("5")},
			},
		},
		ob.GetSnapshot(),
	)
	if err != nil {
		return errors.Trace(err)
	}

	err = ob.ApplyDelta(common.OrderBookDelta{
		SeqNum: 2,
		Bids: common.OrderDeltas{
			Set: []common.PublicOrder{
				common.PublicOrder{Price: dfs("100"), Amount: dfs("8")},
				common.PublicOrder{Price: dfs("96"), Amount: dfs("6")},
				common.PublicOrder{Price: dfs("95"), Amount: dfs("7")},
			},
			Remove: []decimal.Decimal{dfs("92")},
		},
		Asks: common.OrderDeltas{
			Set: []common.PublicOrder{
				common.PublicOrder{Price: dfs("110"), Amount: dfs("9")},
				common.PublicOrder{Price: dfs("130"), Amount: dfs("10")},
			},
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	err = compareSnapshots(
		common.OrderBookSnapshot{
			SeqNum: 2,
			Bids: []common.PublicOrder{
				common.PublicOrder{Price: dfs("100"), Amount: dfs("8")},
				common.PublicOrder{Price: dfs("96"), Amount: dfs("6")},
				common.PublicOrder{Price: dfs("95"), Amount: dfs("7")},
				common.PublicOrder{Price: dfs("91"), Amount: dfs("2")},
				common.PublicOrder{Price: dfs("90"), Amount: dfs("3")},
			},
			Asks: []common.PublicOrder{
				common.PublicOrder{Price: dfs("110"), Amount: dfs("9")},
				common.PublicOrder{Price: dfs("120"), Amount: dfs("5")},
				common.PublicOrder{Price: dfs("130"), Amount: dfs("10")},
			},
		},
		ob.GetSnapshot(),
	)
	if err != nil {
		return errors.Trace(err)
	}

	// Try to apply out-of-band delta: should result in an error
	err = ob.ApplyDelta(common.OrderBookDelta{
		SeqNum: 4,
	})
	if errors.Cause(err) != common.ErrSeqNumMismatch {
		return errors.Errorf("expected ErrSeqNumMismatch, got %v", err)
	}

	ob.ApplySnapshot(common.OrderBookSnapshot{
		SeqNum: 100,
		Bids: []common.PublicOrder{
			common.PublicOrder{Price: dfs("100"), Amount: dfs("1")},
		},
		Asks: []common.PublicOrder{
			common.PublicOrder{Price: dfs("110"), Amount: dfs("2")},
		},
	})

	err = compareSnapshots(
		common.OrderBookSnapshot{
			SeqNum: 100,
			Bids: []common.PublicOrder{
				common.PublicOrder{Price: dfs("100"), Amount: dfs("1")},
			},
			Asks: []common.PublicOrder{
				common.PublicOrder{Price: dfs("110"), Amount: dfs("2")},
			},
		},
		ob.GetSnapshot(),
	)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func compareSnapshots(want, got common.OrderBookSnapshot) error {
	if got.SeqNum != want.SeqNum {
		return errors.Errorf(
			"wanted orderbook update with seqnum %v, got one with %v",
			want.SeqNum, got.SeqNum,
		)
	}

	if err := compareOrders(want.Asks, got.Asks); err != nil {
		return errors.Annotatef(err, "orderbook update with seqnum %v, asks", want.SeqNum)
	}

	if err := compareOrders(want.Bids, got.Bids); err != nil {
		return errors.Annotatef(err, "orderbook update with seqnum %v, bids", want.SeqNum)
	}

	return nil
}

func compareOrders(want, got []common.PublicOrder) error {
	if want == nil {
		want = []common.PublicOrder{}
	}

	if got == nil {
		got = []common.PublicOrder{}
	}

	wantStr := fmt.Sprintf("%+v", want)
	gotStr := fmt.Sprintf("%+v", got)

	if wantStr != gotStr {
		return errors.Errorf("want %s, got %s", wantStr, gotStr)
	}

	return nil
}

// dfs is a shortcut for dfs
func dfs(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}
