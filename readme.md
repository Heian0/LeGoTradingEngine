# LeGoTradingEngine
## Intro
- This project is a high-performance stock exchange simulator designed to model market dynamics and facilitates interaction with order books. It can be used to realistic market condition simulations and test latency-critical workflows, leveraging a Go-based backend for low-latency order handling and a Python-based frontend for real-time visualization of order book states and metrics.
- The simulator can be run as a distributed application (built on gRPC). A server can be used to host the exchange and orderbooks for various securities, and seperate entities can be used send orders to simulate market conditions (via GBM, jump-diffusion, etc). Clients can then connect to the exchange to receive market data as orderbooks the client subscribes to change, allowing trading algorithm logic can be tested via the client.
- I made this project to deepen my understanding about how markets and stock exchanges operate, and to learn more about low-latency distrubuted systems and how to develop them - this project however is far from latency standards required of real world HFT firms using FPGAs and custom message brokers/transfer protocols. When the exchange holds a single orderbook, the orderbook/order matching system is capable of handling approximately 1 million orders per second with a round trip latency of about 1 ms (where orders are siimulated using Geometric Brownian Motion).
- I have previously implemented an orderbook system in C++, but I built this project using Go because I thought it'd be a good opportunity to strenthen my skills in the language.
## OrderBook Implementation
- This exchange supports the following types of orders: Market, Limit, Stop, Trailing Stop.
- It uses an ordered map (implemented with a Red-Black Tree) for price levels, allowing the system to efficiently determine the best bid/ask levels.
- Lookups, insertion, and deletion all run in logarithmic time.
- One latency related con is that iterating through the ordered map is a sequential operation, which can be slower compared to data structures optimized for cache locality like a sorted vector/array.
  # Live Example

https://github.com/user-attachments/assets/c32e0ffd-e74d-426d-95d0-8fefe8707e89

Below is the response and the response time for the Python based GBM market simulator sending 1000000 orders per second to recieve the market state after sending an order.

![image](https://github.com/user-attachments/assets/ec17017d-a110-47bf-9ee0-b938e3d928f9)

This is the client which receives market data that is packaged by the exchange. The time shown is how long it has taken for this update to reach the client since subscribing to the orderbook (should be in line with how long the it has been since the orderbook was subscribed to in real time)

![image](https://github.com/user-attachments/assets/08d06b56-7288-4b0c-81c9-1ed89456fb6c)
## Distributed System Design Overview
- The exchange server can be configured to support any amount of orderbooks, which are tagged with an id and ticker name. Then, simuulators can be used to send orders to the server, and clients can subscribe to specific securities to recieve market data.
- Market update system monitors latency status for connected clients and notfies upon high latency - example below for a high latency client (each client is also given a unique id).
- Sending market data also prioritizes the most recent update to prevent stale data that may build up in the buffer

![image](https://github.com/user-attachments/assets/f435e713-f6fc-4a4c-806a-74fdf65f3435)

A very simple system design diagram.

![image](https://github.com/user-attachments/assets/ee4e0cb4-0876-4e0b-bca0-4143d2beeed5)

## Future Improvements and Extensions
- Profiling would be the first thing to explore - would definitely help learn more about areas which may affect latency in these systems, especially with larger amounts of orderbooks and clients.
- Extending on the idea of multiple orderbooks and clients, a concurrent implementation with a thread pool may greatly lower latency.
- Try for arbitrage - spin up two exchnages which price simulations made to have slight price discrepencies, and try to capitialize on arbitrage positions (read in market data -> realize arbitrage opportunity -> send orders -> mark round trip time).
  




