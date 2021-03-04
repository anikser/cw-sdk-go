package websocket

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	"github.com/shopspring/decimal"

	"code.cryptowat.ch/cw-sdk-go/client/cw"
	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/common"
	"code.cryptowat.ch/cw-sdk-go/config"
	pbb "code.cryptowat.ch/cw-sdk-go/proto/public/broker"
	pbs "code.cryptowat.ch/cw-sdk-go/proto/public/stream"
)

const (
	requestTimeout  = 10 * time.Second
	tradeCacheLimit = 1000
)

// The following errors are returned from TradeClient.
var (
	// ErrNotInitialized is returned when PlaceOrder or CancelOrder is called before the client is initialized.
	// This indicates you have not waited until OnReady was called, or the client is in a disconnected state.
	ErrNotInitialized = errors.New("trade client not initialized")

	// ErrInvalidOrder is returned from PlaceOrder if the order is not valid. Not valid in this case
	// means the order was likely not placed correctly on the exchange.
	ErrInvalidOrder = errors.New("order is not valid")

	// ErrNoExchangeAccess is returned from the OnError callback when the Cryptowatch trading back end
	// does not have proper permissions to access the relevant exchange's API. Usually this means you
	// need to add your API keys to the relevant exchange at https://cryptowat.ch/account/api-keys.
	ErrNoExchangeAccess = errors.New("cryptowatch account is missing exchange api keys")

	// ErrBadProto is returned from PlaceOrder or CancelOrder if the Cryptowatch trading back end returns
	// invalid protobuf data. This should never happen.
	ErrBadProto = errors.New("request is not a valid proto type")
)

type sessionStatusUpdateCB func(marketID common.MarketID, update *pbb.SessionStatusUpdate)

// OrdersUpdateCB defines a callback function for OnOrdersUpdate.
type OrdersUpdateCB func(marketID common.MarketID, orders []common.PrivateOrder)

// PrivateTradesUpdateCB defines a callback function for OnTradesUpdate.
type PrivateTradesUpdateCB func(marketID common.MarketID, trades []common.PrivateTrade)

// OnPositionsUpdateCB defines a callback function for OnPositionsUpdate.
type OnPositionsUpdateCB func(marketID common.MarketID, positions []common.PrivatePosition)

// OnBalancesUpdateCB defines a callback function for OnBalancesUpdate.
type OnBalancesUpdateCB func(marketID common.MarketID, balances []common.Balance)

// OnMarketReadyCB defines a callback function for OnMarketReadyChange.
type OnMarketReadyCB func(marketID common.MarketID, ready bool)

// OnTradeErrorCB defines a callback function for OnError.
// If the error is specific to a market, then marketID is set; otherwise it's
// an empty string. If the error is going to cause the disconnection,
// disconnecting is set to true. In this case, the error listeners are always
// called before the state listeners, so applications can just save the error,
// and display it later, when the disconnection actually happens.
type OnTradeErrorCB func(marketID common.MarketID, err error, disconnecting bool)

// TradeSessionParams is the parameter type for starting a trade session. A trade client connection
// can consist of multiple trade sessions, each of which can have separate exchange auth.
type TradeSessionParams struct {
	// common.MarketParams are parameters to get the market to trade on. You can provide a
	// string-based symbol (exchange+base+quote) or market ID.
	MarketParams common.MarketParams
	// Auth is used as credentials for the exchange. It is optional. If left nil, exchange
	// credentials on your Cryptowatch account will be used.
	ExchangeAuth *ExchangeAuth
}

// Exchange auth is credentials to trade on a particular exchange. CustomerID is specific to
// BitStamp, and KeyPassphrase is specific to coinbase-pro, and can be left blank otherwise.
type ExchangeAuth struct {
	// All exchanges have some form of API key/secret pair.
	APIKey    string
	APISecret string
	// BitStamp
	CustomerID string
	// Coinbase-pro
	KeyPassphrase string
}

type tradeSubscription struct {
	marketID common.MarketID
	auth     *ExchangeAuth
}

// GetResource implements Subscription, so we can use this with the internal ws client
func (ts tradeSubscription) GetResource() string {
	return ts.marketID.String()
}

type callOrdersListenersReq struct {
	marketID  common.MarketID
	update    []common.PrivateOrder
	listeners []OrdersUpdateCB
}

type callTradesListenersReq struct {
	marketID  common.MarketID
	update    []common.PrivateTrade
	listeners []PrivateTradesUpdateCB
}

// Balances grep flag: Ki49fK
type callBalancesListenersReq struct {
	marketID  common.MarketID
	update    []common.Balance
	listeners []OnBalancesUpdateCB
}

type callPositionsListenersReq struct {
	marketID  common.MarketID
	update    []common.PrivatePosition
	listeners []OnPositionsUpdateCB
}

type callSessionStatusListenersReq struct {
	marketID  common.MarketID
	update    *pbb.SessionStatusUpdate
	listeners []sessionStatusUpdateCB
}

type callMarketReadyListenersReq struct {
	marketID  common.MarketID
	ready     bool
	listeners []OnMarketReadyCB
}

type callErrorListenersReq struct {
	marketID      common.MarketID
	err           error
	disconnecting bool
	listeners     []OnTradeErrorCB
}

type placeOrderResp struct {
	order common.PrivateOrder
	err   error
}

type placeOrderReq struct {
	marketID common.MarketID
	params   common.PlaceOrderParams
	response chan placeOrderResp
}

type cancelOrderReq struct {
	marketID common.MarketID
	params   common.CancelOrderParams
	response chan error
}

type syncReq struct {
	marketID common.MarketID
	response chan error
}

type brokerRequest struct {
	id       string
	kind     reqType
	start    time.Time
	response chan *pbb.BrokerUpdateMessage
}

// TradeClient is used to manage a connection to Cryptowatch's trading back end, and
// provides functions for trading on the the subscribed markets. TradeClient
// also has callbacks for updates related to trading.
type TradeClient struct {
	sessionManager *tradeSessionManager

	// Internal cache used for resolving responses to specific ws messages
	requests map[string]*brokerRequest

	// This is the current state of the session which matches the broker backend
	// all values are kept up to date internally.
	orders    map[common.MarketID]map[string]common.PrivateOrder
	trades    map[common.MarketID][]common.PrivateTrade
	positions map[common.MarketID]map[string]common.PrivatePosition
	balances  common.Balances

	placeOrderRequests  chan placeOrderReq
	cancelOrderRequests chan cancelOrderReq
	syncRequests        chan syncReq

	callOrdersListeners             chan callOrdersListenersReq
	callTradesListeners             chan callTradesListenersReq
	callBalancesListeners           chan callBalancesListenersReq
	callPositionsListeners          chan callPositionsListenersReq
	callSessionStatusListeners      chan callSessionStatusListenersReq
	callSubscriptionResultListeners chan callSubscriptionResultListenersReq
	callErrorListeners              chan callErrorListenersReq

	ordersListeners             []OrdersUpdateCB
	tradesListeners             []PrivateTradesUpdateCB
	balancesListeners           []OnBalancesUpdateCB
	positionsListeners          []OnPositionsUpdateCB
	marketReadyListeners        []OnMarketReadyCB
	sessionStatusListeners      []sessionStatusUpdateCB
	subscriptionResultListeners []SubscriptionResultCB

	onReadyCallbacks []func()
	onErrorCallbacks []OnTradeErrorCB

	disable chan struct{}

	// We want to ensure that wsConn's methods aren't available on the
	// TradeClient to avoid confusion, so we give it explicit name.
	wsConn *wsConn

	cwClient cw.Interface
	mtx      sync.Mutex
}

type TradeClientParams struct {
	WSParams      *WSParams
	RESTURL       string
	TradeSessions []*TradeSessionParams
}

// NewTradeClient creates a new TradeClient based on the given WSParams. Subscriptions
// should be an array of market IDs that you have access to trade on.
func NewTradeClient(params *TradeClientParams) (*TradeClient, error) {
	rc := rest.NewRESTClient(&rest.RESTClientParams{
		URL: params.RESTURL,
	})
	cwClient := cw.NewCWClient(&cw.CWClientParams{
		RESTClient: rc,
	})
	return newTradeClientInternal(params, cwClient)
}

func newTradeClientInternal(params *TradeClientParams, cwClient cw.Interface) (
	*TradeClient, error,
) {
	// The trade client can't take a nil config currengly
	if params == nil {
		return nil, errors.New("missing trade client params")
	}

	// Make a copy of params struct because we might alter it below
	paramsCopy := *params
	params = &paramsCopy

	// Fall back on config in ~/.cw/credentials.yml
	var cfg = config.Get()

	if params.WSParams == nil {
		params.WSParams = &WSParams{}
	}

	if params.WSParams.APIKey == "" && params.WSParams.SecretKey == "" {
		params.WSParams.APIKey = cfg.APIKey
		params.WSParams.SecretKey = cfg.SecretKey
	}

	if params.WSParams.URL == "" {
		params.WSParams.URL = cfg.TradeURL
	}

	if params.RESTURL == "" {
		params.RESTURL = cfg.RESTURL
	}

	sessionManager, err := newTradeSessionManager(params.TradeSessions, cwClient)
	if err != nil {
		return nil, errors.Trace(err)
	}

	wsConn, err := newWsConn(
		params.WSParams,
		&wsConnParamsInternal{
			unmarshalAuthnResult: unmarshalAuthnResultTrade,
			subscriptions:        tradeSubsToSubs(sessionManager.getSubscriptions()),
		},
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	tc := &TradeClient{
		wsConn:   wsConn,
		cwClient: cwClient,

		// Used to signal shutdown from a lost connection
		disable:        make(chan struct{}, 1),
		sessionManager: sessionManager,

		// Internal request cache
		requests: make(map[string]*brokerRequest),

		orders:    make(map[common.MarketID]map[string]common.PrivateOrder),
		trades:    make(map[common.MarketID][]common.PrivateTrade),
		positions: make(map[common.MarketID]map[string]common.PrivatePosition),
		balances:  make(map[common.Exchange][]common.Balance),

		placeOrderRequests:  make(chan placeOrderReq, 1),
		cancelOrderRequests: make(chan cancelOrderReq, 1),
		syncRequests:        make(chan syncReq, 1),

		callOrdersListeners:             make(chan callOrdersListenersReq, 1),
		callTradesListeners:             make(chan callTradesListenersReq, 1),
		callBalancesListeners:           make(chan callBalancesListenersReq, 1),
		callPositionsListeners:          make(chan callPositionsListenersReq, 1),
		callSessionStatusListeners:      make(chan callSessionStatusListenersReq, 1),
		callSubscriptionResultListeners: make(chan callSubscriptionResultListenersReq, 1),
		callErrorListeners:              make(chan callErrorListenersReq, 1),
	}

	tc.wsConn.onRead(func(data []byte) {
		var msg pbb.BrokerUpdateMessage

		if err := proto.Unmarshal(data, &msg); err != nil {
			// Failed to parse incoming message: close connection (and if
			// reconnection was requested, then reconnect)
			tc.wsConn.disconnectOpt(nil, websocket.CloseUnsupportedData, "", false)
			return
		}

		switch msg.Update.(type) {
		// Exposed
		case *pbb.BrokerUpdateMessage_OrdersUpdate:
			tc.ordersUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetOrdersUpdate())

		case *pbb.BrokerUpdateMessage_TradesUpdate:
			tc.tradesUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetTradesUpdate())

		case *pbb.BrokerUpdateMessage_BalancesUpdate:
			tc.balancesUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetBalancesUpdate())

		case *pbb.BrokerUpdateMessage_PositionsUpdate:
			tc.positionsUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetPositionsUpdate())

		case *pbb.BrokerUpdateMessage_SubscriptionResult:
			tc.subscriptionResultHandler(msg.GetSubscriptionResult())

		case *pbb.BrokerUpdateMessage_SessionStatusUpdate:
			tc.sessionStatusUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetSessionStatusUpdate())

		// Internal order resolution
		case *pbb.BrokerUpdateMessage_RequestResolutionUpdate:
			var id string
			switch m := msg.Update.(type) {
			case *pbb.BrokerUpdateMessage_RequestResolutionUpdate:
				id = m.RequestResolutionUpdate.Id
			}

			if id == "" {
				return
			}

			tc.mtx.Lock()
			bReq, ok := tc.requests[id]
			tc.mtx.Unlock()

			if !ok {
				return
			}

			if time.Now().Sub(bReq.start) > requestTimeout {
				close(bReq.response)
				return
			}

			tc.requestWithIDHandler(id, &msg)

		case *pbb.BrokerUpdateMessage_ApiAccessorStatusUpdate:
			// We're interested in this update to figure whether we have access to
			// the exchange.
			tc.apiAccessorStatusUpdateHandler(common.MarketID(msg.GetMarketId()), msg.GetApiAccessorStatusUpdate())

		case *pbb.BrokerUpdateMessage_PermissionsUpdate:
			// TODO there may be error states in here we should handle
			// fmt.Println(msg.GetPermissionsUpdate())

		case *pbb.BrokerUpdateMessage_AnonymousSessionStatusUpdate:
			// Not used

		default:
			// Not a supported type
		}
	})

	tc.wsConn.onError(func(err error, disconnecting bool) {
		tc.scheduleCallErrorHandlers(0, errors.Trace(err), disconnecting)
	})

	tc.wsConn.onConnClosed(func(_ ConnState) {
		tc.disable <- struct{}{}
	})

	go tc.listen()

	return tc, nil

}

// This serves as the main program logic. It runs in its own goroutine and
// handles order requests as well as dispatching listener callbacks. This
// design asllows the client to place/cancel as many orders at a time that
// they want while retaining a synchronous api.
func (tc *TradeClient) listen() {
listenloop:
	for {
		select {
		case req := <-tc.placeOrderRequests:
			if !tc.sessionManager.sessionReady(req.marketID) {
				req.response <- placeOrderResp{
					err: ErrNotInitialized,
				}
				continue
			}

			go func() {
				req.response <- tc.placeOrderInt(req.params)
			}()

		case req := <-tc.cancelOrderRequests:
			if !tc.sessionManager.sessionReady(req.marketID) {
				req.response <- ErrNotInitialized
				continue
			}

			go func() {
				req.response <- tc.cancelOrderInt(req.params)
			}()

		case req := <-tc.syncRequests:
			go func() {
				req.response <- tc.syncInt(req.marketID)
			}()

		case req := <-tc.callOrdersListeners:
			tc.mtx.Lock()
			tc.sessionManager.setModuleReady(req.marketID, tradeModuleOrders)
			tc.mtx.Unlock()
			for _, l := range req.listeners {
				l(req.marketID, req.update)
			}

		case req := <-tc.callSubscriptionResultListeners:
			for _, l := range req.listeners {
				l(req.result)
			}

		case req := <-tc.callTradesListeners:
			tc.mtx.Lock()
			tc.sessionManager.setModuleReady(req.marketID, tradeModuleTrades)
			tc.mtx.Unlock()
			for _, l := range req.listeners {
				l(req.marketID, req.update)
			}

		case req := <-tc.callBalancesListeners:
			tc.mtx.Lock()
			tc.sessionManager.setModuleReady(req.marketID, tradeModuleBalances)
			tc.mtx.Unlock()
			for _, l := range req.listeners {
				l(req.marketID, req.update)
			}

		case req := <-tc.callPositionsListeners:
			tc.mtx.Lock()
			tc.sessionManager.setModuleReady(req.marketID, tradeModulePositions)
			tc.mtx.Unlock()
			for _, l := range req.listeners {
				l(req.marketID, req.update)
			}

		case req := <-tc.callSessionStatusListeners:
			tc.mtx.Lock()
			prev := tc.sessionManager.setModuleReady(req.marketID, tradeModuleInitialized)
			tc.mtx.Unlock()
			for _, l := range req.listeners {
				l(req.marketID, req.update)
			}

			// When we receive the first session status update, and if the session
			// was synced more than 5 mins ago, request the resync.
			if !prev {
				lastSyncTime := time.Unix(req.update.LastSyncTime, 0)
				if time.Since(lastSyncTime) > 5*time.Minute {
					go tc.syncInt(req.marketID)
				}
			}

		case req := <-tc.callErrorListeners:
			for _, l := range req.listeners {
				l(req.marketID, req.err, req.disconnecting)
			}

		case <-tc.disable:
			tc.mtx.Lock()
			tc.sessionManager.reset()
			tc.mtx.Unlock()
			continue listenloop
		}

		// Call market ready listeners once
		for _, session := range tc.sessionManager.getSessions() {
			session.onReadyOnce(func() {
				tc.mtx.Lock()
				cbs := make([]OnMarketReadyCB, len(tc.marketReadyListeners))
				copy(cbs, tc.marketReadyListeners)
				tc.mtx.Unlock()

				for _, cb := range cbs {
					cb(session.marketID, true)
				}
			})
		}

		// Call OnReady callbacks once
		tc.sessionManager.onReadyOnce(func() {
			cbs := []func(){}
			tc.mtx.Lock()
			cbs = append(cbs, tc.onReadyCallbacks...)
			tc.mtx.Unlock()
			for _, cb := range cbs {
				cb()
			}
		})
	}
}

type reqType string

func (r reqType) String() string {
	return string(r)
}

const (
	reqPlaceOrder  reqType = "place-order"
	reqCancelOrder reqType = "cancel-order"
	reqSync        reqType = "sync"
)

func (tc *TradeClient) makeRequest(kind reqType, req *pbb.BrokerRequest) (*pbb.BrokerUpdateMessage, error) {
	bReq := &brokerRequest{
		id:       uuid.New().String(),
		kind:     kind,
		response: make(chan *pbb.BrokerUpdateMessage, 1),
	}

	req.Id = bReq.id

	tc.mtx.Lock()
	bReq.start = time.Now()
	tc.requests[bReq.id] = bReq
	tc.mtx.Unlock()

	defer func() {
		tc.mtx.Lock()
		delete(tc.requests, bReq.id)
		tc.mtx.Unlock()
	}()

	if err := tc.wsConn.sendProto(context.Background(), req); err != nil {
		return nil, errors.Annotatef(err, "tradeclient %s request failed", bReq.kind)
	}

	select {
	case r, ok := <-bReq.response:
		if !ok {
			return nil, errors.Errorf("tradeclient %s request failed", bReq.kind)
		}

		return r, nil
	case <-time.After(requestTimeout):
		return nil, errors.Errorf("tradeclient %s request failed: timed out after %s", bReq.kind, requestTimeout.String())
	}
}

func (tc *TradeClient) apiAccessorStatusUpdateHandler(marketID common.MarketID, res *pbb.APIAccessorStatusUpdate) {
	if !res.HasAccess {
		// Cryptowatch account doesn't have access to the exchange, so report that
		// to the user.
		//
		// TODO: when we have multiple sessions per connection, create a dedicated
		// type for that kind of error, so that clients can figure which exchange
		// we don't have access to.

		tc.scheduleCallErrorHandlers(marketID, ErrNoExchangeAccess, false)
	}
}

func (tc *TradeClient) subscriptionResultHandler(update *pbs.SubscriptionResult) {
	result := subscriptionResultFromProto(update)

	tc.mtx.Lock()
	subresListeners := make([]SubscriptionResultCB, len(tc.subscriptionResultListeners))
	copy(subresListeners, tc.subscriptionResultListeners)
	tc.mtx.Unlock()

	tc.callSubscriptionResultListeners <- callSubscriptionResultListenersReq{
		result:    result,
		listeners: subresListeners,
	}
}

func (tc *TradeClient) scheduleCallErrorHandlers(
	marketID common.MarketID, err error, disconnecting bool,
) {
	tc.mtx.Lock()
	cbs := make([]OnTradeErrorCB, len(tc.onErrorCallbacks))
	copy(cbs, tc.onErrorCallbacks)
	tc.mtx.Unlock()

	tc.callErrorListeners <- callErrorListenersReq{
		marketID:      marketID,
		err:           err,
		listeners:     cbs,
		disconnecting: disconnecting,
	}
}

func (tc *TradeClient) requestWithIDHandler(id string, msg *pbb.BrokerUpdateMessage) {
	tc.mtx.Lock()
	bReq, ok := tc.requests[id]
	tc.mtx.Unlock()

	if !ok {
		// This means the server double-sent a request resolution update. This should
		// never happen, but we can ignore anyway
		return
	}

	bReq.response <- msg
	close(bReq.response)
}

// PlaceOrder creates a new order based on the given OrderParams. PlaceOrder blocks
// until the order has been placed on the exchange or an error occurs. PlaceOrder
// can be called concurrently as many times as needed.
func (tc *TradeClient) PlaceOrder(params common.PlaceOrderParams) (common.PrivateOrder, error) {
	response := make(chan placeOrderResp, 1)

	tc.placeOrderRequests <- placeOrderReq{
		marketID: params.MarketID,
		params:   params,
		response: response,
	}

	res := <-response

	return res.order, res.err
}

func (tc *TradeClient) placeOrderInt(params common.PlaceOrderParams) placeOrderResp {
	failedOrder := placeOrderResp{
		order: common.PrivateOrder{Error: 500},
	}

	resp, err := tc.makeRequest(reqPlaceOrder, &pbb.BrokerRequest{
		MarketId: int64(params.MarketID),
		Request: &pbb.BrokerRequest_PlaceOrderRequest{
			PlaceOrderRequest: &pbb.PlaceOrderRequest{
				Order: placeOrderParamsToProto(params),
			},
		},
	})

	if err != nil {
		failedOrder.err = errors.Annotatef(err, "request failed: %s", reqPlaceOrder)
		return failedOrder
	}

	result := resp.GetRequestResolutionUpdate()
	if result == nil {
		failedOrder.err = errors.Errorf("request failed: %s", reqPlaceOrder)
		return failedOrder
	}

	if result.Error != 0 {
		return placeOrderResp{
			order: common.PrivateOrder{Error: result.Error},
			err:   errors.Errorf("[%v] %v", result.Error, result.Message),
		}
	}

	poRes := result.GetPlaceOrderResult()
	if poRes == nil {
		failedOrder.err = errors.Errorf("request failed: %s", reqPlaceOrder)
		return failedOrder
	}

	order := privateOrderFromProto(poRes.Order)

	tc.mtx.Lock()
	tc.orders[params.MarketID][order.CacheKey(params.MarketID)] = order
	tc.mtx.Unlock()

	return placeOrderResp{
		order: order,
	}
}

// CancelOrder cancels the given order on the exchange. CancelOrder blocks
// until the order has been placed or if an error occurs. it can be called
// concurrently on as many different orders as needed.
func (tc *TradeClient) CancelOrder(params common.CancelOrderParams) error {
	if params.OrderID == "" {
		return ErrInvalidOrder
	}

	response := make(chan error, 1)

	tc.cancelOrderRequests <- cancelOrderReq{
		marketID: params.MarketID,
		params:   params,
		response: response,
	}

	return <-response
}

func (tc *TradeClient) cancelOrderInt(params common.CancelOrderParams) error {
	resp, err := tc.makeRequest(reqCancelOrder, &pbb.BrokerRequest{
		MarketId: int64(params.MarketID),
		Request: &pbb.BrokerRequest_CancelOrderRequest{
			CancelOrderRequest: &pbb.CancelOrderRequest{
				OrderId: params.OrderID,
			},
		},
	})

	if err != nil {
		return errors.Annotatef(err, "request failed: %s", reqCancelOrder)
	}

	result := resp.GetRequestResolutionUpdate()
	if result == nil {
		return errors.Annotatef(err, "request failed: %s", reqCancelOrder)
	}

	if result.Error != 0 {
		return errors.Errorf("[%v] %v", result.Error, result.Message)
	}

	return nil
}

// Sync forces a cache update by polling the exchange on behalf of the user.
// This function should not normally be needed, and is only useful in two scenarios:
// 1) an order is placed or cancelled outside of this client.
// 2) there is something preventing our trading back end from actively polling for updates.
// This happens rarely, and for various reasons. For example, an exchange may rate limit
// one of our servers.
func (tc *TradeClient) Sync(marketID common.MarketID) error {
	response := make(chan error, 1)

	tc.syncRequests <- syncReq{
		marketID: marketID,
		response: response,
	}

	return <-response
}

func (tc *TradeClient) syncInt(marketID common.MarketID) error {
	resp, err := tc.makeRequest(reqSync, &pbb.BrokerRequest{
		MarketId: int64(marketID),
		Request: &pbb.BrokerRequest_SyncRequest{
			SyncRequest: &pbb.SyncRequest{},
		},
	})

	if err != nil {
		return errors.Annotatef(err, "request failed: %s", reqSync)
	}

	result := resp.GetRequestResolutionUpdate()
	if result == nil {
		return errors.Annotatef(err, "request failed: %s", reqSync)
	}

	if result.Error != 0 {
		return errors.Errorf("[%v] %v", result.Error, result.Message)
	}

	return nil
}

// GetOrders returns the list of current open orders ordered by execution time
// (oldest first). If the market is not yet ready, returns ErrNotInitialized.
func (tc *TradeClient) GetOrders(marketID common.MarketID) ([]common.PrivateOrder, error) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	if !tc.sessionManager.sessionReady(marketID) {
		return nil, ErrNotInitialized
	}

	var ods []common.PrivateOrder
	for _, order := range tc.orders[marketID] {
		ods = append(ods, order)
	}

	sort.Sort(common.PrivateOrders(ods))

	return ods, nil
}

// GetTrades returns the 1000 most recent trades ordered by execution time
// (oldest first). If the market is not yet ready, returns ErrNotInitialized.
func (tc *TradeClient) GetTrades(marketID common.MarketID) ([]common.PrivateTrade, error) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	if !tc.sessionManager.sessionReady(marketID) {
		return nil, ErrNotInitialized
	}

	tds := make([]common.PrivateTrade, len(tc.trades[marketID]))
	copy(tds, tc.trades[marketID])

	sort.Sort(common.PrivateTrades(tds))

	return tds, nil
}

// GetPositions returns the list of open positions ordered by execution time
// (oldest first). If the market is not yet ready, returns ErrNotInitialized.
func (tc *TradeClient) GetPositions(marketID common.MarketID) ([]common.PrivatePosition, error) {
	if !tc.sessionManager.sessionReady(marketID) {
		return nil, ErrNotInitialized
	}

	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	var ps []common.PrivatePosition
	for _, position := range tc.positions[marketID] {
		ps = append(ps, position)
	}

	sort.Sort(common.PrivatePositions(ps))

	return ps, nil
}

func (tc *TradeClient) GetBalances() (map[common.Exchange][]common.Balance, error) {
	if !tc.sessionManager.allSessionsReady() {
		return nil, ErrNotInitialized
	}

	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	balances := make(common.Balances, len(tc.balances))
	for xch, xchBalances := range tc.balances {
		xchBalancesCpy := make([]common.Balance, len(xchBalances))
		copy(xchBalancesCpy, xchBalances)
		balances[xch] = xchBalancesCpy
	}

	return balances, nil
}

func (tc *TradeClient) ordersUpdateHandler(marketID common.MarketID, update *pbb.OrdersUpdate) {
	orders := update.GetOrders()
	var orderCache = make(map[string]common.PrivateOrder, len(orders))
	var returnOrders []common.PrivateOrder

	for _, order := range orders {
		o := privateOrderFromProto(order)
		orderCache[o.CacheKey(marketID)] = o
		returnOrders = append(returnOrders, o)
	}

	tc.mtx.Lock()

	// Update internal order cache and call order update listeners
	tc.orders[marketID] = orderCache

	// Handle listeners
	listeners := make([]OrdersUpdateCB, len(tc.ordersListeners))
	copy(listeners, tc.ordersListeners)

	tc.mtx.Unlock()

	tc.callOrdersListeners <- callOrdersListenersReq{
		marketID:  marketID,
		update:    returnOrders,
		listeners: listeners,
	}
}

func (tc *TradeClient) tradesUpdateHandler(marketID common.MarketID, update *pbb.TradesUpdate) {
	trades := update.GetTrades()
	var newTrades = make([]common.PrivateTrade, 0, len(trades))

	for _, trade := range trades {
		newTrades = append(newTrades, tradeFromProto(trade))
	}

	tc.mtx.Lock()

	tradeSlice := tc.trades[marketID]
	tradeSlice = append(tradeSlice, newTrades...)

	// Make sure the trade cache doesn't exceed tradeCacheLimit by removing
	// oldest trades.
	if len(tradeSlice) > tradeCacheLimit {
		tradeSlice = tradeSlice[(len(tradeSlice) - tradeCacheLimit):]
	}

	tc.trades[marketID] = tradeSlice

	listeners := make([]PrivateTradesUpdateCB, len(tc.tradesListeners))
	copy(listeners, tc.tradesListeners)

	tc.mtx.Unlock()

	tc.callTradesListeners <- callTradesListenersReq{
		marketID:  marketID,
		update:    newTrades,
		listeners: listeners,
	}
}

func (tc *TradeClient) balancesUpdateHandler(marketID common.MarketID, update *pbb.BalancesUpdate) {
	var balances []common.Balance

	for _, balancesByFType := range update.GetBalances() {
		fType := balancesByFType.FundingType

		for _, b := range balancesByFType.Balances {
			balances = append(balances, common.Balance{
				FundingType: common.FundingType(fType),
				Asset: tc.cwClient.MustGetAsset(cw.GetAssetParams{
					Symbol: common.AssetSymbol(b.Currency),
				}),
				Amount: decimal.RequireFromString(b.AmountString),
			})
		}
	}

	market := tc.cwClient.MustGetMarket(cw.GetMarketParams{
		ID: marketID,
	})

	tc.mtx.Lock()
	tc.balances[market.Exchange] = balances

	listeners := make([]OnBalancesUpdateCB, len(tc.balancesListeners))
	copy(listeners, tc.balancesListeners)

	tc.mtx.Unlock()

	tc.callBalancesListeners <- callBalancesListenersReq{
		marketID:  marketID,
		update:    balances,
		listeners: listeners,
	}
}

func (tc *TradeClient) positionsUpdateHandler(marketID common.MarketID, update *pbb.PositionsUpdate) {
	positions := update.GetPositions()
	var positionsCache = make(map[string]common.PrivatePosition, len(positions))
	var returnPositions []common.PrivatePosition

	for _, position := range positions {
		p := positionFromProto(position)
		positionsCache[p.ExternalID] = p
		returnPositions = append(returnPositions, p)
	}

	tc.mtx.Lock()

	tc.positions[marketID] = positionsCache

	listeners := make([]OnPositionsUpdateCB, len(tc.positionsListeners))
	copy(listeners, tc.positionsListeners)

	tc.mtx.Unlock()

	tc.callPositionsListeners <- callPositionsListenersReq{
		marketID:  marketID,
		update:    returnPositions,
		listeners: listeners,
	}
}

func (tc *TradeClient) sessionStatusUpdateHandler(marketID common.MarketID, update *pbb.SessionStatusUpdate) {
	tc.mtx.Lock()
	listeners := make([]sessionStatusUpdateCB, len(tc.sessionStatusListeners))
	copy(listeners, tc.sessionStatusListeners)
	tc.mtx.Unlock()

	tc.callSessionStatusListeners <- callSessionStatusListenersReq{
		marketID:  marketID,
		update:    update,
		listeners: listeners,
	}
}

// OnReady sets a callback function for when the client is fully initialized and ready to place/cancel trades.
// PlaceOrder and CancelOrder will return ErrNotInitialized if called before OnReady is called. OnReady is called
// each time the client initializes, meaning if the client disconnects for whatever reason, and then reconnects,
// it will be called again. Any number of OnReady callbacks can be placed, and can be done so safely from multiple
// goroutines.
func (tc *TradeClient) OnReady(cb func()) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.onReadyCallbacks = append(tc.onReadyCallbacks, cb)
}

// OnError sets a callback function for errors associated with the trading
// client or trading back end.
func (tc *TradeClient) OnError(cb OnTradeErrorCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.onErrorCallbacks = append(tc.onErrorCallbacks, cb)
}

// OnSubscriptionResult is called whenever a subscription attempt was made; it
// happens after the connection and authentication is successful.
func (tc *TradeClient) OnSubscriptionResult(cb SubscriptionResultCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.subscriptionResultListeners = append(tc.subscriptionResultListeners, cb)
}

// OnOrdersUpdate sets a callback for order updates. This will be called
// immediately when the client initializes, and for every subsequent order
// update. Each time this callback is executed, the entire list of open orders
// is returned (including partially filled orders).
func (tc *TradeClient) OnOrdersUpdate(cb OrdersUpdateCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.ordersListeners = append(tc.ordersListeners, cb)
}

// OnTradesUpdate sets a callabck for trade updates. This will be called
// immediately when the client initializes with the 1000 most recent trades.
// For every subsequent update, it will contain only the most recent trades.
func (tc *TradeClient) OnTradesUpdate(cb PrivateTradesUpdateCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.tradesListeners = append(tc.tradesListeners, cb)
}

// Balances grep flag: Ki49fK
// OnBalancesUpdate sets a callback for balance updates. This will be called
// immediately when the client initializes, and for every subsequent balance
// update. An internal cache of balances is kept, and can be accessed with
// GetBalances()
func (tc *TradeClient) OnBalancesUpdate(cb OnBalancesUpdateCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.balancesListeners = append(tc.balancesListeners, cb)
}

// OnPositionsUpdate sets a callback for position updates. This will be
// called immediately when the client initializes, and for every subsequent
// position update. An internal cache of positions is kept, and can be accessed
// with GetPositions().
func (tc *TradeClient) OnPositionsUpdate(cb OnPositionsUpdateCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.positionsListeners = append(tc.positionsListeners, cb)
}

// OnMarketReadyChange registers a callback which is called when the market
// ready status changes, i.e. when the market becomes ready or not ready for
// placing orders.
func (tc *TradeClient) OnMarketReadyChange(cb OnMarketReadyCB) {
	tc.mtx.Lock()
	defer tc.mtx.Unlock()

	tc.marketReadyListeners = append(tc.marketReadyListeners, cb)
}

// OnStateChange registers a new listener for the given state. The listener is
// registered with the default options (call the listener every time the state
// becomes active, and don't call the listener immediately for the current
// state). All registered callbacks for all states (and all messages, see
// OnMarketData) will be called by the same internal goroutine, i.e. they are
// never called concurrently with each other.
//
// The order of listeners invocation for the same state is unspecified, and
// clients shouldn't rely on it.
//
// The listeners shouldn't block; a blocked listener will also block the whole
// stream connection.
//
// To subscribe to all state changes, use ConnStateAny as a state.
func (tc *TradeClient) OnStateChange(state ConnState, cb StateCallback) {
	tc.wsConn.onStateChange(state, cb)
}

// OnStateChangeOpt is like OnStateChange, but also takes additional
// options; see StateListenerOpt for details.
func (tc *TradeClient) OnStateChangeOpt(state ConnState, cb StateCallback, opt StateListenerOpt) {
	tc.wsConn.onStateChangeOpt(state, cb, opt)
}

// GetSubscriptions returns a slice of the current subscriptions.
func (tc *TradeClient) GetSubscriptions() []*tradeSession {
	return tc.sessionManager.getSessions()
}

// OnConnClosed allows the client to set a callback for when the connection is lost.
// The new state of the client could be ConnStateDisconnected or ConnStateWaitBeforeReconnect.
func (tc *TradeClient) OnConnClosed(cb ConnClosedCallback) {
	tc.wsConn.onConnClosed(cb)
}

// URL returns the url the client is connected to, e.g. wss://stream.cryptowat.ch.
func (tc *TradeClient) URL() string {
	return tc.wsConn.url()
}

// Connect either starts a connection goroutine (if state is
// ConnStateDisconnected), or makes it connect immediately, ignoring timeout
// (if the state is ConnStateWaitBeforeReconnect). For other states, this returns an
// error.
//
// Connect doesn't wait for the connection to establish; it returns immediately.
func (tc *TradeClient) Connect() (err error) {
	return tc.wsConn.connect()
}

// Close stops the connection (or reconnection loop, if active), and if
// websocket connection is active at the moment, closes it as well.
func (tc *TradeClient) Close() (err error) {
	return tc.wsConn.close()
}

func tradeSubsToSubs(tradeSubs []*tradeSubscription) []Subscription {
	subs := make([]Subscription, 0, len(tradeSubs))

	for _, v := range tradeSubs {
		subs = append(subs, v)
	}

	return subs
}

func subsToTradeSubs(subs []Subscription) []*tradeSubscription {
	tradeSubs := make([]*tradeSubscription, 0, len(subs))

	for _, sub := range subs {
		v, ok := sub.(*tradeSubscription)
		if !ok {
			panic(errInvalidSubType)
		}

		tradeSubs = append(tradeSubs, v)
	}

	return tradeSubs
}
