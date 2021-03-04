package common

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetDeltasAgainst(t *testing.T) {
	type testCase struct {
		descr       string
		snapshotOld OrderBookSnapshot
		snapshotNew OrderBookSnapshot
		wantDelta   OrderBookDelta
	}

	testCases := []testCase{
		testCase{descr: "Equal snapshots", // {{{
			snapshotOld: OrderBookSnapshot{
				SeqNum: 1,
				Asks: []PublicOrder{
					PublicOrder{
						Price:  dfs("10.0"),
						Amount: dfs("10.1"),
					},
					PublicOrder{
						Price:  dfs("11.0"),
						Amount: dfs("11.1"),
					},
					PublicOrder{
						Price:  dfs("12.0"),
						Amount: dfs("12.1"),
					},
				},
				Bids: []PublicOrder{
					PublicOrder{
						Price:  dfs("9.0"),
						Amount: dfs("9.1"),
					},
					PublicOrder{
						Price:  dfs("8.0"),
						Amount: dfs("8.1"),
					},
					PublicOrder{
						Price:  dfs("7.0"),
						Amount: dfs("7.1"),
					},
				},
			},
			snapshotNew: OrderBookSnapshot{
				SeqNum: 1,
				Asks: []PublicOrder{
					PublicOrder{
						Price:  dfs("10.0"),
						Amount: dfs("10.1"),
					},
					PublicOrder{
						Price:  dfs("11.0"),
						Amount: dfs("11.1"),
					},
					PublicOrder{
						Price:  dfs("12.0"),
						Amount: dfs("12.1"),
					},
				},
				Bids: []PublicOrder{
					PublicOrder{
						Price:  dfs("9.0"),
						Amount: dfs("9.1"),
					},
					PublicOrder{
						Price:  dfs("8.0"),
						Amount: dfs("8.1"),
					},
					PublicOrder{
						Price:  dfs("7.0"),
						Amount: dfs("7.1"),
					},
				},
			},
			wantDelta: OrderBookDelta{
				SeqNum: 1,
			},
		},
		// }}}
		testCase{descr: "2 new orders, 1 order to remove (separately)", // {{{
			snapshotOld: OrderBookSnapshot{
				SeqNum: 1,
				Asks: []PublicOrder{
					PublicOrder{Price: dfs("11.0"), Amount: dfs("11.1")},
				},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("9.1")},
					PublicOrder{Price: dfs("8.0"), Amount: dfs("8.1")},
					PublicOrder{Price: dfs("7.0"), Amount: dfs("7.1")},
				},
			},
			snapshotNew: OrderBookSnapshot{
				SeqNum: 1,
				Asks: []PublicOrder{
					PublicOrder{Price: dfs("10.0"), Amount: dfs("10.1")},
					PublicOrder{Price: dfs("11.0"), Amount: dfs("11.1")},
					PublicOrder{Price: dfs("12.0"), Amount: dfs("12.1")},
				},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("9.1")},
					PublicOrder{Price: dfs("8.0"), Amount: dfs("8.1")},
				},
			},
			wantDelta: OrderBookDelta{
				SeqNum: 1,
				Asks: OrderDeltas{
					Set: []PublicOrder{
						PublicOrder{Price: dfs("10.0"), Amount: dfs("10.1")},
						PublicOrder{Price: dfs("12.0"), Amount: dfs("12.1")},
					},
				},
				Bids: OrderDeltas{
					Remove: []decimal.Decimal{
						dfs("7.0"),
					},
				},
			},
		},
		// }}}
		testCase{descr: "2 orders to update amounts", // {{{
			snapshotOld: OrderBookSnapshot{
				SeqNum: 1,
				Asks: []PublicOrder{
					PublicOrder{Price: dfs("11.0"), Amount: dfs("11.1")},
				},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("9.1")},
					PublicOrder{Price: dfs("8.0"), Amount: dfs("8.1")},
					PublicOrder{Price: dfs("7.0"), Amount: dfs("7.1")},
				},
			},
			snapshotNew: OrderBookSnapshot{
				SeqNum: 2,
				Asks: []PublicOrder{
					PublicOrder{Price: dfs("11.0"), Amount: dfs("11.1")},
				},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("90.1")},
					PublicOrder{Price: dfs("8.0"), Amount: dfs("80.1")},
					PublicOrder{Price: dfs("7.0"), Amount: dfs("7.1")},
				},
			},
			wantDelta: OrderBookDelta{
				SeqNum: 2,
				Bids: OrderDeltas{
					Set: []PublicOrder{
						PublicOrder{Price: dfs("9.0"), Amount: dfs("90.1")},
						PublicOrder{Price: dfs("8.0"), Amount: dfs("80.1")},
					},
				},
			},
		},
		// }}}
		testCase{descr: "all types of delta together", // {{{
			snapshotOld: OrderBookSnapshot{
				SeqNum: 1,
				Asks:   []PublicOrder{},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("9.1")},
					PublicOrder{Price: dfs("8.0"), Amount: dfs("8.1")},
					PublicOrder{Price: dfs("7.0"), Amount: dfs("7.1")},
					PublicOrder{Price: dfs("6.0"), Amount: dfs("6.1")},
					PublicOrder{Price: dfs("5.0"), Amount: dfs("5.1")},
				},
			},
			snapshotNew: OrderBookSnapshot{
				SeqNum: 2,
				Asks:   []PublicOrder{},
				Bids: []PublicOrder{
					PublicOrder{Price: dfs("9.0"), Amount: dfs("9.1")},
					PublicOrder{Price: dfs("6.0"), Amount: dfs("6.1")},
					PublicOrder{Price: dfs("5.0"), Amount: dfs("50.1")},
					PublicOrder{Price: dfs("4.0"), Amount: dfs("4.1")},
					PublicOrder{Price: dfs("3.0"), Amount: dfs("3.1")},
					PublicOrder{Price: dfs("2.0"), Amount: dfs("2.1")},
				},
			},
			wantDelta: OrderBookDelta{
				SeqNum: 2,
				Bids: OrderDeltas{
					Set: []PublicOrder{
						PublicOrder{Price: dfs("5.0"), Amount: dfs("50.1")},
						PublicOrder{Price: dfs("4.0"), Amount: dfs("4.1")},
						PublicOrder{Price: dfs("3.0"), Amount: dfs("3.1")},
						PublicOrder{Price: dfs("2.0"), Amount: dfs("2.1")},
					},
					Remove: []decimal.Decimal{
						dfs("8.0"),
						dfs("7.0"),
					},
				},
			},
		},
		// }}}
	}

	for i, tc := range testCases {
		gotDelta := tc.snapshotNew.GetDeltasAgainst(tc.snapshotOld)
		assert.Equal(t, tc.wantDelta, gotDelta, "test case #%d (%s)", i, tc.descr)
	}
}

// TestDeltas tests both ApplyDeltas and GetDeltasAgainst.
func TestDeltas(t *testing.T) {
	// For every test case the following is done:
	// - Apply delta to ordersOld, ensure that the result is equal to
	//   wantOrdersNew
	// - Regenerate delta from ordersOld and wantOrdersNew, and make sure it has
	//   the same effect as the provided delta.
	// - Regenerate delta in the opposite direction (from wantOrdersNew and
	//   wantOrdersNew), and make sure that applying it to the new orders yields
	//   old orders.
	// - Reverse ordersOld and wantOrdersNew, and do the same 3 things as above
	//   but with the reversed flag.
	type testCase struct {
		descr string
		// ordersOld is a ASC-sorted (like asks) slice of orders to which the delta
		// will be applied.
		//
		// If it's nil, then the last snapshot after last applied delta will be
		// used.
		//
		// NOTE that test case should have it sorted in ascending order (like
		// asks); but it will be tested in both ascending and descending orders.
		ordersOld []PublicOrder
		// delta is the delta to apply to ordersOld
		delta OrderDeltas
		// wantOrdersNew is the expected orders ASC-sorted. It will be tested in
		// both ascending and descending orders.
		wantOrdersNew []PublicOrder
	}

	orders4 := []PublicOrder{
		PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
		PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
		PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
		PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
	}

	orders11 := []PublicOrder{
		PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
		PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
		PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
		PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
		PublicOrder{Price: dfs("94"), Amount: dfs("9.4")},
		PublicOrder{Price: dfs("95"), Amount: dfs("9.5")},
		PublicOrder{Price: dfs("96"), Amount: dfs("9.6")},
		PublicOrder{Price: dfs("97"), Amount: dfs("9.7")},
		PublicOrder{Price: dfs("98"), Amount: dfs("9.8")},
		PublicOrder{Price: dfs("99"), Amount: dfs("9.9")},
		PublicOrder{Price: dfs("100"), Amount: dfs("10.0")},
	}

	testCases := []testCase{
		testCase{descr: "Insert in the beginning", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Insert in the middle", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("92.2"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("92.2"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Insert in the end", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("93.2"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
				PublicOrder{Price: dfs("93.2"), Amount: dfs("1000")},
			},
		}, // }}}
		testCase{descr: "Insert in the beginning, middle and the end", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("93.2"), Amount: dfs("1002")},
					PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("92.2"), Amount: dfs("1001")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("92.2"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
				PublicOrder{Price: dfs("93.2"), Amount: dfs("1002")},
			},
		}, // }}}
		testCase{descr: "Insert multiple in the beginning, middle and the end", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("89.0"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("89.1"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("92.2"), Amount: dfs("1002")},
					PublicOrder{Price: dfs("92.3"), Amount: dfs("1003")},
					PublicOrder{Price: dfs("93.2"), Amount: dfs("1004")},
					PublicOrder{Price: dfs("93.5"), Amount: dfs("1005")},
					PublicOrder{Price: dfs("93.6"), Amount: dfs("1006")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("89.0"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("89.1"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("92.2"), Amount: dfs("1002")},
				PublicOrder{Price: dfs("92.3"), Amount: dfs("1003")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
				PublicOrder{Price: dfs("93.2"), Amount: dfs("1004")},
				PublicOrder{Price: dfs("93.5"), Amount: dfs("1005")},
				PublicOrder{Price: dfs("93.6"), Amount: dfs("1006")},
			},
		}, // }}}

		testCase{descr: "Replace in the beginning", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Replace in the middle", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("92"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Replace in the end", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("93"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("1000")},
			},
		}, // }}}

		testCase{descr: "2 redundant replacements of the same order: last one wins", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "2 redundant replacements of the same order: last one wins", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "4 redundant replacements of the same order: last one wins; also a bunch of other deltas", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("20"), Amount: dfs("100")},
					PublicOrder{Price: dfs("92.1"), Amount: dfs("200")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("92.2"), Amount: dfs("201")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1002")},
					PublicOrder{Price: dfs("19"), Amount: dfs("101")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("19"), Amount: dfs("101")},
				PublicOrder{Price: dfs("20"), Amount: dfs("100")},
				PublicOrder{Price: dfs("90"), Amount: dfs("1002")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("92.1"), Amount: dfs("200")},
				PublicOrder{Price: dfs("92.2"), Amount: dfs("201")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}

		testCase{descr: "2 redundant inserts of the same order: last one wins", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90.1"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("90.1"), Amount: dfs("1001")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("90.1"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "2 redundant inserts of the same order: last one wins", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90.1"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("90.1"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("90.1"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}

		testCase{descr: "Insert and replace in the beginning", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("89"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Replace and insert in the beginning", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("90.1"), Amount: dfs("1000")},
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("90.1"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}

		testCase{descr: "Delete from the beginning", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Remove: []decimal.Decimal{dfs("90")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Delete from the middle", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Remove: []decimal.Decimal{dfs("92")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Delete from the end", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Remove: []decimal.Decimal{dfs("93")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
			},
		}, // }}}
		testCase{descr: "Delete everything", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Remove: []decimal.Decimal{
					dfs("90"), dfs("91"), dfs("92"), dfs("93"),
				},
			},
			wantOrdersNew: []PublicOrder{},
		}, // }}}
		testCase{descr: "Delete everything and insert new ones", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("88"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("89"), Amount: dfs("1001")},
				},
				Remove: []decimal.Decimal{
					dfs("90"), dfs("91"), dfs("92"), dfs("93"),
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("88"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("89"), Amount: dfs("1001")},
			},
		}, // }}}
		testCase{descr: "Delete everything and insert new ones, but some of them are old orders: they are still deleted", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("88"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("90"), Amount: dfs("1000")},
					PublicOrder{Price: dfs("89"), Amount: dfs("1001")},
					PublicOrder{Price: dfs("91"), Amount: dfs("1000")},
				},
				Remove: []decimal.Decimal{
					dfs("90"), dfs("91"), dfs("92"), dfs("93"),
				},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("88"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("89"), Amount: dfs("1001")},
			},
		}, // }}}

		testCase{descr: "Remove non-existing order: no-op", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Remove: []decimal.Decimal{dfs("91.1")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}
		testCase{descr: "Insert and remove the same order: no-op", // {{{
			ordersOld: orders4,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("91.1"), Amount: dfs("1001")},
				},
				Remove: []decimal.Decimal{dfs("91.1")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
			},
		}, // }}}

		testCase{descr: "Insert, replace, remove", // {{{
			ordersOld: orders11,
			delta: OrderDeltas{
				Set: []PublicOrder{
					PublicOrder{Price: dfs("92.1"), Amount: dfs("1000")}, // Insert
					PublicOrder{Price: dfs("95.2"), Amount: dfs("1001")}, // Insert
					PublicOrder{Price: dfs("96"), Amount: dfs("1002")},   // Replace
					PublicOrder{Price: dfs("97"), Amount: dfs("1003")},   // Replace
				},
				Remove: []decimal.Decimal{dfs("99")},
			},
			wantOrdersNew: []PublicOrder{
				PublicOrder{Price: dfs("90"), Amount: dfs("9.0")},
				PublicOrder{Price: dfs("91"), Amount: dfs("9.1")},
				PublicOrder{Price: dfs("92"), Amount: dfs("9.2")},
				PublicOrder{Price: dfs("92.1"), Amount: dfs("1000")},
				PublicOrder{Price: dfs("93"), Amount: dfs("9.3")},
				PublicOrder{Price: dfs("94"), Amount: dfs("9.4")},
				PublicOrder{Price: dfs("95"), Amount: dfs("9.5")},
				PublicOrder{Price: dfs("95.2"), Amount: dfs("1001")},
				PublicOrder{Price: dfs("96"), Amount: dfs("1002")},
				PublicOrder{Price: dfs("97"), Amount: dfs("1003")},
				PublicOrder{Price: dfs("98"), Amount: dfs("9.8")},
				PublicOrder{Price: dfs("100"), Amount: dfs("10.0")},
			},
		}, // }}}
	}

	lastOrders := []PublicOrder{}
	for i, tc := range testCases {
		if tc.ordersOld == nil {
			tc.ordersOld = lastOrders
		}

		// Test ascending order // {{{
		testSnapshotsAndDelta(t, tc.ordersOld, tc.wantOrdersNew, tc.delta, false, i, tc.descr)
		// }}}

		// Test descending order {{{
		ordersOldReversed := make([]PublicOrder, len(tc.ordersOld))
		for i, o := range tc.ordersOld {
			ordersOldReversed[len(ordersOldReversed)-1-i] = o
		}

		wantOrdersNewReversed := make([]PublicOrder, len(tc.wantOrdersNew))
		for i, o := range tc.wantOrdersNew {
			wantOrdersNewReversed[len(wantOrdersNewReversed)-1-i] = o
		}

		testSnapshotsAndDelta(t, ordersOldReversed, wantOrdersNewReversed, tc.delta, true, i, tc.descr)
		// }}}
	}
}

func testSnapshotsAndDelta(
	t *testing.T, ordersOld, wantOrdersNew []PublicOrder, delta OrderDeltas,
	reversed bool, tcNum int, tcDescr string,
) {
	ordersNew := ApplyDeltas(ordersOld, &delta, reversed)
	assert.Equal(t, wantOrdersNew, ordersNew, "%stest case #%d (%s)", reversedStr(reversed), tcNum, tcDescr)

	// Regenerate delta and ensure that it has the same effect as original delta.
	// NOTE: we can't easily compare calcualtedDelta and delta, because delta
	// could have some redundancy. So the easiest way to ensure they have the
	// same effect is to reapply calcualtedDelta to ordersOld and ensure that the
	// resulting snapshot is the same as ordersNew.

	calculatedDelta := GetDeltasAgainst(ordersOld, ordersNew, reversed)
	ordersNew2 := ApplyDeltas(ordersOld, &calculatedDelta, reversed)
	assert.Equal(t, wantOrdersNew, ordersNew2, "%sAFTER REGENERATED DELTA test case #%d (%s)", reversedStr(reversed), tcNum, tcDescr)

	// Also regenerate delta in the opposite direction (from ordersNew to
	// ordersOld), and make sure it's correct.
	calculatedDeltaOpposite := GetDeltasAgainst(ordersNew, ordersOld, reversed)
	ordersOld2 := ApplyDeltas(ordersNew, &calculatedDeltaOpposite, reversed)
	assert.Equal(t, ordersOld, ordersOld2, "%sAFTER REGENERATED OPPOSITE DELTA test case #%d (%s)", reversedStr(reversed), tcNum, tcDescr)
}

func reversedStr(reversed bool) string {
	if !reversed {
		return ""
	}

	return "REVERSED "
}

// dfs is a shortcut for decimal.RequireFromString
func dfs(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}
