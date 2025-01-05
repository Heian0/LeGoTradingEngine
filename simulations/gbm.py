import numpy as np
import yfinance as yf
import matplotlib.pyplot as plt
from datetime import datetime, timedelta
from dataclasses import dataclass
from typing import List
from exchange_pb2 import OrderMessage, Command, OrderResponseMessage, OrderTimeInForce, OrderType, Side
from exchange_pb2_grpc import ExchangeServiceStub
import sys
import time
import grpc

class SPYSimulator:
    def __init__(self):
        # Get historical data to estimate parameters
        spy = yf.Ticker('SPY')
        hist = spy.history(period='2y')
        prices = hist['Close']
        
        # Calculate log returns
        log_returns = np.log(prices / prices.shift(1)).dropna()
        
        # Calculate daily parameters
        self.daily_vol = np.std(log_returns)
        self.daily_drift = np.mean(log_returns)
        
        # Store last price as starting point
        self.current_price = prices.iloc[-1]
        
        # Print estimated parameters
        print(f"Starting price: ${self.current_price:.2f}")
        print(f"Annual volatility: {self.daily_vol * np.sqrt(252):.2%}")
        print(f"Annual drift: {self.daily_drift * 252:.2%}")
    
    def simulate_path(self, days=1, steps_per_day=390):
        """
        Simulate price path
        days: number of days to simulate
        steps_per_day: number of steps per day (390 = number of minutes in trading day)
        """
        total_steps = days * steps_per_day
        dt = 1/(252 * steps_per_day)  # Time step in years
        
        # Generate price path
        prices = np.zeros(total_steps)
        prices[0] = self.current_price
        
        # Random walks
        z = np.random.normal(0, 1, total_steps)
        
        # Generate path using GBM
        for t in range(1, total_steps):
            drift = (self.daily_drift - 0.5 * self.daily_vol**2) * dt
            diffusion = self.daily_vol * np.sqrt(dt) * z[t]
            prices[t] = prices[t-1] * np.exp(drift + diffusion)
        
        return prices

    def plot_simulation(self, n_paths=5):
        """Plot multiple simulation paths"""
        plt.figure(figsize=(12, 6))
        
        for _ in range(n_paths):
            prices = self.simulate_path()
            plt.plot(prices)
            
        plt.title(f'SPY Price Simulation ({n_paths} paths)')
        plt.xlabel('Minutes')
        plt.ylabel('Price')
        plt.grid(True)
        plt.savefig('simulation_plot.png')
        plt.close()

class RealtimeOrderGenerator:
    def __init__(self, price_simulator: 'SPYSimulator'):
        self.simulator = price_simulator
        self.minutes_per_day = 390
        self.orders_per_second = 2
        self.interval = 1.0 / self.orders_per_second  # Time between orders

        # Setup gRPC connection
        self.channel = grpc.insecure_channel('localhost:9000')
        self.stub = ExchangeServiceStub(self.channel)
        print("Connected to order service at localhost:9000")
        
    def run_continuously(self):
        try:
            while True:
                fair_prices = self.simulator.simulate_path(steps_per_day=self.minutes_per_day)
                current_minute = 0
                curr_id = 0
                
                while current_minute < self.minutes_per_day:
                    fair_price = fair_prices[current_minute]
                    start_time = time.time()
                    
                    while time.time() - start_time < 60:
                        timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
                        order = self._generate_single_limit_order(fair_price, curr_id, 0)
                        curr_id += 1
                        
                        # Send order via gRPC
                        try:
                            response = self.stub.HandleOrder(order)
                            side = "BUY" if order.orderSide == Side.BID else "SELL"
                            price_dollars = order.price / 100
                            print(response.exchangeStatus)
                            #print(f"{timestamp} - {side} {order.quantity} @ {price_dollars:.2f}")
                        except grpc.RpcError as e:
                            print(f"Failed to send order: {e}")
                            
                        time.sleep(self.interval)
                        
                    current_minute += 1
                
                print(f"{datetime.now().strftime('%H:%M:%S')},NEXTDAY")
                
        except KeyboardInterrupt:
            print("\nClosing gRPC connection...")
            self.channel.close()
            sys.exit(0)
    
    def _generate_single_limit_order(self, fair_price: float, curr_id: int, symbol_id: int):
        """Generate a single limit order"""
        # Convert price to cents
        fair_price_cents = int(fair_price * 100)
        
        # Decide buy/sell
        is_buy = np.random.random() < 0.5
        
        # Generate price offset using exponential distribution
        # More orders near fair price, fewer far away
        mean_offset = 10  # 10 cents average offset
        offset = int(np.random.exponential(mean_offset))
        
        if is_buy:
            price = fair_price_cents - offset  # Buyers bid below fair price
        else:
            price = fair_price_cents + offset  # Sellers ask above fair price
            
        # Generate size using power law distribution
        # Many small orders, few large orders
        min_size = 100  # Minimum order size
        alpha = 1.5    # Power law exponent
        size = int(min_size / np.power(np.random.random(), 1/alpha))
        
        # Round size to nearest lot of 100
        size = (size // 100) * 100
        if size < 100:
            size = 100
            
        order = OrderMessage()
        order.command = Command.ADD
        order.orderType = OrderType.LIMIT
        order.orderTimeInForce = OrderTimeInForce.GTC
        if is_buy:
            order.orderSide = Side.BID
        else:
            order.orderSide = Side.ASK
        order.id = curr_id
        order.symbolId = symbol_id
        order.price = price
        order.stopPrice = 0
        order.trailingAmount = 0
        order.lastExecutedPrice = 0
        order.quantity = size
        order.openQuantity = 0
        order.lastExecutedQuantity = 0
        return order

if __name__ == "__main__":
    # Initialize simulator
    sim = SPYSimulator()
    
    # Generate and plot multiple paths
    sim.plot_simulation()
    
    # Generate single path for minute-by-minute prices
    prices = sim.simulate_path()
    
    # Print some statistics
    print(f"\nSimulation statistics:")
    print(f"Min price: ${min(prices):.2f}")
    print(f"Max price: ${max(prices):.2f}")
    print(f"Mean price: ${np.mean(prices):.2f}")
    print(f"Final price: ${prices[-1]:.2f}")

    """
    # Initialize simulator and generator
    sim = SPYSimulator()
    generator = OrderGenerator(sim)
    
    # Generate orders for one day
    orders = generator.generate_limit_orders()
    
    # Print first 10 orders
    print("\nFirst 10 orders:")
    for order in orders[:10]:
        side = "BUY" if order.orderSide == Side.BID else "SELL"
        price_dollars = order.price / 100
        print(f"{side}: {order.quantity} shares @ ${price_dollars:.2f}")
        
    # Print some statistics
    print(f"\nTotal orders generated: {len(orders)}")
    print(f"Average order size: {np.mean([o.quantity for o in orders]):.0f} shares")
    buy_orders = [o for o in orders if o.orderSide == Side.BID]
    sell_orders = [o for o in orders if o.orderSide == Side.ASK]
    print(f"Buy orders: {len(buy_orders)}")
    print(f"Sell orders: {len(sell_orders)}")
    """

    sim = SPYSimulator()
    generator = RealtimeOrderGenerator(sim)
    generator.run_continuously()