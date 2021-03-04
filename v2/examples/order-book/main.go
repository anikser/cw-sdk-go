package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"code.cryptowat.ch/cw-sdk-go/client/rest"
	"code.cryptowat.ch/cw-sdk-go/client/websocket"
	"code.cryptowat.ch/cw-sdk-go/common"
	"code.cryptowat.ch/cw-sdk-go/orderbooks"
)

func main() {
	restClient := rest.NewRESTClient(nil)
	streamClient, err := websocket.NewStreamClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	streamClient.Connect()

	orderbook, err := orderbooks.NewOrderBookWatcher(orderbooks.OrderBookWatcherParams{
		RESTClient:   restClient,
		StreamClient: streamClient,
		Market: common.MarketParams{
			Symbol: common.MarketSymbol{
				Exchange: "kraken",
				Base:     "btc",
				Quote:    "usd",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	orderbook.OnUpdate(func(update orderbooks.Update) {
		if ob := update.OrderBookUpdate; ob != nil {
			log.Println("Order book update", ob)
		} else if err := update.GetSnapshotError; err != nil {
			log.Println("Error", err)
		}
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// Wait until the OS signal is received, at which point we'll close the connection and quit.
	<-signals
	log.Print("Exiting")
}
