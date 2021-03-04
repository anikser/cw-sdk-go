package websocket

import (
	"fmt"
	"sync"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/client/cw"
	"code.cryptowat.ch/cw-sdk-go/common"
)

// tradeSession is used to keep track of what modules are ready per-session.
// a session is not considered initialized until it has received all the necessary
// updates like positions, balances, trades, etc. These are called tradeModule.
type tradeSession struct {
	marketID common.MarketID
	modules  map[tradeModule]bool
	ready    sync.Once
}

func (ts *tradeSession) String() string {
	ret := "trade session: market=" + ts.marketID.String()
	for _, m := range tradeModules {
		ret += fmt.Sprintf(" %s=%v", m, ts.isModuleReady(m))
	}
	return ret
}

func (tsm *tradeSessionManager) newTradeSession(opt *TradeSessionParams) error {
	// Need to resolve the market, which makes sure the market/exchange/assets are cached and
	// everything will work
	m, err := tsm.cwClient.GetMarket(cw.GetMarketParams(opt.MarketParams))
	if err != nil {
		return errors.Trace(err)
	}

	tsm.sessions[m.ID] = &tradeSession{
		marketID: m.ID,
		modules:  make(map[tradeModule]bool, len(tradeModules)),
		ready:    sync.Once{},
	}

	return nil
}

func (ts *tradeSession) setModuleReady(m tradeModule) {
	ts.modules[m] = true
}

func (ts *tradeSession) isModuleReady(m tradeModule) bool {
	return ts.modules[m]
}

func (ts *tradeSession) allModulesReady() bool {
	for _, module := range tradeModules {
		if !ts.isModuleReady(module) {
			return false
		}
	}
	return true
}

func (ts *tradeSession) onReadyOnce(f func()) {
	if ts.allModulesReady() {
		ts.ready.Do(f)
	}
}

type tradeSessionManager struct {
	sessions map[common.MarketID]*tradeSession
	cwClient cw.Interface
	ready    sync.Once
}

func newTradeSessionManager(opts []*TradeSessionParams, cwClient cw.Interface) (
	*tradeSessionManager, error,
) {
	tsm := &tradeSessionManager{
		sessions: make(map[common.MarketID]*tradeSession),
		cwClient: cwClient,
	}

	for _, sessionOpts := range opts {
		err := tsm.newTradeSession(sessionOpts)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return tsm, nil
}

func (tsm *tradeSessionManager) getSessions() []*tradeSession {
	ret := make([]*tradeSession, 0, len(tsm.sessions))
	for _, s := range tsm.sessions {
		ret = append(ret, s)
	}
	return ret
}

func (tsm *tradeSessionManager) getSubscriptions() []*tradeSubscription {
	var subscriptions []*tradeSubscription
	for _, s := range tsm.getSessions() {
		subscriptions = append(subscriptions, &tradeSubscription{
			marketID: s.marketID,
		})
	}
	return subscriptions
}

func (tsm *tradeSessionManager) reset() {
	for _, session := range tsm.getSessions() {
		err := tsm.newTradeSession(&TradeSessionParams{
			MarketParams: common.MarketParams{
				ID: session.marketID,
			},
		})
		// If err isn't nil, that means the cache is broken, since the objects
		// should already be resolved.
		if err != nil {
			panic(err)
		}
	}

	tsm.ready = sync.Once{}
}

func (tsm *tradeSessionManager) sessionReady(m common.MarketID) bool {
	return tsm.sessions[m].allModulesReady()
}

func (tsm *tradeSessionManager) allSessionsReady() bool {
	for _, session := range tsm.sessions {
		if !session.allModulesReady() {
			return false
		}
	}
	return true
}

func (tsm *tradeSessionManager) onReadyOnce(f func()) {
	if tsm.allSessionsReady() {
		tsm.ready.Do(f)
	}
}

// setModuleReady sets the new module ready status, and returns the previous one.
func (tsm *tradeSessionManager) setModuleReady(marketID common.MarketID, module tradeModule) bool {
	ret := tsm.sessions[marketID].isModuleReady(module)
	tsm.sessions[marketID].setModuleReady(module)
	return ret
}

func (tsm *tradeSessionManager) isModuleReady(marketID common.MarketID, module tradeModule) bool {
	return tsm.sessions[marketID].isModuleReady(module)
}

type tradeModule string

const (
	tradeModuleInitialized tradeModule = "initialized"
	tradeModuleOrders      tradeModule = "orders"
	tradeModuleTrades      tradeModule = "trades"
	tradeModulePositions   tradeModule = "positions"
	tradeModuleBalances    tradeModule = "balances"
)

var tradeModules = [...]tradeModule{
	tradeModuleInitialized,
	tradeModuleOrders,
	tradeModuleTrades,
	tradeModulePositions,
	tradeModuleBalances,
}
