# LeGoTradingEngine
## Intro
- This project is a high-performance stock exchange simulator designed to model market dynamics and facilitates interaction with order books. It can be used to realistic market condition simulations and test latency-critical workflows, leveraging a Go-based backend for low-latency order handling and a Python-based frontend for real-time visualization of order book states and metrics.
- The simulator can be run as a distributed application (built on gRPC). A server can be used to host the exchange and orderbooks for various securities, and seperate entities can be used send orders to simulate market conditions (via GBM, jump-diffusion, etc). Clients can then connect to the exchange to receive market data as orderbooks the client subscribes to change, allowing trading algorithm logic can be tested via the client.
- I made this project to deepen my understanding about how markets and stock exchanges operate, and to learn more about low-latency distrubuted systems and how to develop them - this project however is far from latency standards required of real world HFT firms using FPGAs and custom message brokers/transfer protocols. When the exchange holds a single orderbook, the orderbook/order matching system is capable of handling approximately 1 million orders per second with a round trip latency of about 1 ms (where orders are siimulated using Geometric Brownian Motion).
- I have previously implemented an orderbook system in C++, but I built this project using Go because I thought it'd be a good opportunity to strenthen my skills in the language.
# OrderBook Implementation
- This exchange supports the following types of orders: Market, Limit, Stop, Trailing Stop.
- It uses an ordered map (implemented with a Red-Black Tree) for price levels, allowing the system to efficiently determine the best bid/ask levels.
- Lookups, insertion, and deletion all run in logarithmic time.
- One latency related con is that iterating through the ordered map is a sequential operation, which can be slower compared to data structures optimized for cache locality like a sorted vector/array.








