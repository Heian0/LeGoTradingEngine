package main

import (
	"log"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
)

type Strategy struct {
	//mdaID uint64
	mda *BasicMarketDataAggregator
}

func NewStrategy(mda *BasicMarketDataAggregator) *Strategy {
	return &Strategy{
		//mdaID: mda.Subscribe(),
		mda: mda,
	}
}

/*
func (s *Strategy) Run() {
	for {
		state, ok := s.mda.ReadUpdate(s.mdaID)
		if !ok {
			time.Sleep(time.Microsecond)
			continue
		}

		// Process update
		s.ProcessUpdate(state)
	}
}
*/

func (s *Strategy) ProcessUpdate(state *exg.OrderBookState) {
	// Your strategy logic here
	log.Printf("Strategy processing update - Best Bid: %v, Best Ask: %v",
		state.BestBid, state.BestAsk)
}
