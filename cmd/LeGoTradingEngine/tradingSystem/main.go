package main

import (
	"fmt"
	"log"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
)

func displayOrderBook(state *exg.OrderBookState) {
	fmt.Print(state.ObsToString())
}

func main() {
	mda, err := NewBasicMarketDataAggregator(8011)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer mda.Close()

	if err := mda.ListenForUpdates(); err != nil {
		log.Fatalf("Subscription error: %v", err)
	}
}
