package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/client/websocket"
	"code.cryptowat.ch/cw-sdk-go/common"
)

func main() {
	restclient := rest.NewRESTClient(nil)

	// Get market descriptions, to know symbols (like "btcusd") by integer ID.
	marketsSlice, err := restclient.GetMarketsIndex()
	if err != nil {
		log.Fatalf("failed to get markets: %s", err)
	}

	markets := map[common.MarketID]rest.MarketDescr{}
	for _, market := range marketsSlice {
		markets[common.MarketID(market.ID)] = market
	}

	// Create a new stream connection instance. Note that the actual connection
	// will happen later
	c, err := websocket.NewStreamClient(&websocket.StreamClientParams{
		WSParams: &websocket.WSParams{
			// The following credentials can also be read from ~/.cw/credentials.yml
			// APIKey:    <your api key>,
			// SecretKey: <your secret key>,
		},

		Subscriptions: []*websocket.StreamSubscription{
			&websocket.StreamSubscription{
				// Subscription key for all trades from all markets
				Resource: "markets:*:trades",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Listen for received trades and print them
	c.OnMarketUpdate(func(marketID common.MarketID, md common.MarketUpdate) {
		if md.TradesUpdate == nil {
			return
		}

		tradesUpdate := md.TradesUpdate
		for _, trade := range tradesUpdate.Trades {
			log.Printf(
				"%s	%s	%-25s	%-25s",
				markets[marketID].Exchange,
				trade.OrderSide,
				fmt.Sprintf("Price: %s", trade.Price),
				fmt.Sprintf("Amount: %s", trade.Amount),
			)
		}
	})

	// Connect to stream
	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}

	// Wait until an OS signal is received, at which point we'll close the connection and quit
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-signals

	log.Print("Closing connection...")

	if err := c.Close(); err != nil {
		log.Fatalf("Failed to close connection: %s", err)
	}
}
