from market_data_aggregator import MarketDataAggregator

class ArbitrageModel:
    def __init__(self):
        self.nyse_stat

if __name__ == "__main__":
    market_data_aggregator = MarketDataAggregator()
    try:
        market_data_aggregator.run()
    except KeyboardInterrupt:
        print("\nShutting down...")