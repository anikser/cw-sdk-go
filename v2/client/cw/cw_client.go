package cw

import (
	"sync"

	"github.com/juju/errors"

	"code.cryptowat.ch/cw-sdk-go/cache"
	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/common"
)

type Interface interface {
	GetMarket(GetMarketParams) (common.Market, error)
	MustGetMarket(GetMarketParams) common.Market
	GetAsset(GetAssetParams) (common.Asset, error)
	MustGetAsset(GetAssetParams) common.Asset
}

// CWClient is a convenience tool for getting Cryptowatch objects from the api.
// Instead of having to create a REST client, CWClient makes one internally
// based on the config settings.
type CWClient struct {
	rc    rest.V2
	cache *cache.Cache
}

// CWClientParams allows you to control the RESTClient in CWClient.
type CWClientParams struct {
	RESTClient rest.V2
}

// NewCWClient creates a new instance of CWClient. It will use the default RESTClient.
func NewCWClient(opt *CWClientParams) *CWClient {
	if opt == nil {
		opt = &CWClientParams{}
	}

	optCopy := *opt
	opt = &optCopy

	if opt.RESTClient == nil {
		opt.RESTClient = rest.NewRESTClient(nil)
	}

	return &CWClient{
		rc:    opt.RESTClient,
		cache: cache.New(),
	}
}

// GetMarket uses the REST API to get a market by exchange + base + quote. This method
// is not thread-safe.
func (cw *CWClient) GetMarket(opt GetMarketParams) (common.Market, error) {
	if opt.ID != 0 {
		if m, hit := cw.cache.GetMarketByID(opt.ID); hit {
			return m, nil
		}

		m, err := cw.rc.GetMarketByID(opt.ID)
		if err != nil {
			return common.Market{}, errors.Trace(err)
		}

		cw.cache.SetMarket(m)
		return m, nil
	}

	if m, hit := cw.cache.GetMarketBySymbol(opt.Symbol); hit {
		return m, nil
	}

	m, err := cw.rc.GetMarketBySymbol(opt.Symbol)
	if err != nil {
		return common.Market{}, errors.Trace(err)
	}

	cw.cache.SetMarket(m)
	return m, nil
}

func (cw *CWClient) MustGetMarket(opt GetMarketParams) common.Market {
	m, err := cw.GetMarket(opt)
	if err != nil {
		panic(err)
	}
	return m
}

func (cw *CWClient) GetAsset(opt GetAssetParams) (common.Asset, error) {
	if opt.ID != 0 {
		if a, hit := cw.cache.GetAssetByID(opt.ID); hit {
			return a, nil
		}

		a, err := cw.rc.GetAssetByID(opt.ID)
		if err != nil {
			return common.Asset{}, errors.Trace(err)
		}

		cw.cache.SetAsset(a)
		return a, nil
	}

	if a, hit := cw.cache.GetAssetBySymbol(opt.Symbol); hit {
		return a, nil
	}

	a, err := cw.rc.GetAssetBySymbol(opt.Symbol)
	if err != nil {
		return common.Asset{}, errors.Trace(err)
	}

	cw.cache.SetAsset(a)
	return a, nil
}

func (cw *CWClient) MustGetAsset(opt GetAssetParams) common.Asset {
	a, err := cw.GetAsset(opt)
	if err != nil {
		panic(err)
	}
	return a
}

var (
	cwSingleton *CWClient
	cwMtx       sync.Mutex
)

// GetMarket is a convenience method on package cw which does not require an instance of CWClient.
func GetMarket(opt GetMarketParams) (common.Market, error) {
	cwMtx.Lock()
	defer cwMtx.Unlock()

	if cwSingleton == nil {
		cwSingleton = NewCWClient(nil)
	}

	return cwSingleton.GetMarket(opt)
}

// GetAsset is a convenience method which does not require an instance of CWClient.
func GetAsset(opt GetAssetParams) (common.Asset, error) {
	cwMtx.Lock()
	defer cwMtx.Unlock()

	if cwSingleton == nil {
		cwSingleton = NewCWClient(nil)
	}

	return cwSingleton.GetAsset(opt)
}
