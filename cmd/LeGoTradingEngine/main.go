package main

func main() {

	/*

		exchange := exg.NewExchange()
		go exchange.InitializeExchangeUI()

		exchange.AddOrderbook(0, "TSLA")

		// Command line input loop
		fmt.Println("Enter new exchange data (or 'quit' to exit):")
		for {
			var input string
			fmt.Scanln(&input)

			if input == "quit" {
				os.Exit(0)
			}

			if input == "ask" {
				newAskLimitOrder1 := ob.LimitAskOrder(1, 0, 100, 30, ob.GoodTillCancel)
				exchange.AddOrder(&newAskLimitOrder1)
				newAskLimitOrder2 := ob.LimitAskOrder(2, 0, 100, 20, ob.GoodTillCancel)
				exchange.AddOrder(&newAskLimitOrder2)
			}

			if input == "bid" {

				newBidLimitOrder1 := ob.LimitBidOrder(3, 0, 100, 5, ob.GoodTillCancel)
				exchange.AddOrder(&newBidLimitOrder1)
				newBidLimitOrder2 := ob.LimitBidOrder(4, 0, 100, 10, ob.GoodTillCancel)
				exchange.AddOrder(&newBidLimitOrder2)
			}

			if input == "ask2" {
				newAskLimitOrder3 := ob.LimitAskOrder(5, 0, 100, 30, ob.GoodTillCancel)
				exchange.AddOrder(&newAskLimitOrder3)
				newAskLimitOrder4 := ob.LimitAskOrder(6, 0, 100, 20, ob.GoodTillCancel)
				exchange.AddOrder(&newAskLimitOrder4)
			}

		}

			exchange := exg.NewExchange()
			fmt.Println(exchange.Name)
			go exchange.Test()
			time.Sleep(2 * time.Second)
			fmt.Println(exchange.Name)

			exchange.PrintExchange()
			exchange.AddOrderbook(0, "TSLA")
			exchange.AddOrderbook(1, "NVDA")
			newAskLimitOrder1 := ob.LimitAskOrder(1, 1, 100, 30, ob.GoodTillCancel)
			exchange.AddOrder(&newAskLimitOrder1)
			exchange.PrintExchange()
	*/

	/*
		ob := orderbook.NewOrderbook(999)

		newAskLimitOrder1 := orderbook.LimitAskOrder(1, 999, 100, 30, orderbook.GoodTillCancel)
		ob.AddOrder(&newAskLimitOrder1)
		fmt.Println(ob.String())

		newBidLimitOrder1 := orderbook.LimitBidOrder(2, 999, 50, 10, orderbook.GoodTillCancel)
		ob.AddOrder(&newBidLimitOrder1)
		fmt.Println(ob.String())

		newBidStopOrder1 := orderbook.StopBidOrder(3, 999, 8, 20, orderbook.GoodTillCancel)
		ob.AddOrder(&newBidStopOrder1)
		fmt.Println(ob.String())

		newAskLimitOrder2 := orderbook.LimitAskOrder(4, 999, 20, 5, orderbook.GoodTillCancel)
		ob.AddOrder(&newAskLimitOrder2)
		fmt.Println(ob.String())
	*/
}
