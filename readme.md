# LeGoTradingEngine
# Intro
- This project is a stock exchange simulator that can be used to simulate market conditions and interact with orderbooks. It is a distibuted application where a server can be used to host orderbooks for various securities, and seperate entities can be used to simulate market conditions. Lastly, clients can connect to the exchange to receive market data as orderbooks the client subscribes to change, allowing trading algorithm logic can be tested via the client.
- I made this project to deepen my understanding about how markets and stock exchanges operate, and to learn more about low-latency distrubuted systems and how to develop them - this project however is far from latency standards requires of HFT. When the exchange holds a single orderbook, the orderbook/order matching system is capable of handling approximately 1 million orders per second with a round trip latency of about 1 ms (where orders are siimulated using Geometric Brownian Motion).
- I have previously implemented an orderbook system in C++, but I built this project using Go because I thought it'd be a good opportunity to strenthen my skills in the language.
