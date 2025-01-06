import sys
import zmq
from PyQt6.QtWidgets import (QApplication, QMainWindow, QTableWidget, 
                            QTableWidgetItem, QVBoxLayout, QHBoxLayout, QWidget, QLabel)
from PyQt6.QtCore import QThread, pyqtSignal, Qt
from PyQt6.QtGui import QColor, QPalette

# Import your generated protobuf code
from exchange_pb2 import OrderBookState  

class OrderBookWorker(QThread):
    update_signal = pyqtSignal(object)

    def __init__(self):
        super().__init__()
        self.context = zmq.Context()
        self.subscriber = self.context.socket(zmq.SUB)
        self.subscriber.connect("tcp://localhost:5555")
        self.subscriber.setsockopt_string(zmq.SUBSCRIBE, "")
        self.subscriber.set_hwm(1000)  # High water mark to prevent memory issues

    def run(self):
        while True:
            try:
                data = self.subscriber.recv()  # Receive raw bytes
                state = OrderBookState()
                state.ParseFromString(data)  # Parse protobuf message
                self.update_signal.emit(state)
            except Exception as e:
                print(f"Error receiving data: {e}")

class StyledTableItem(QTableWidgetItem):
    def __init__(self, text, is_bid=True):
        super().__init__(text)
        self.setTextAlignment(Qt.AlignmentFlag.AlignRight | Qt.AlignmentFlag.AlignVCenter)
        color = QColor("#2ecc71") if is_bid else QColor("#e74c3c")  # Green for bids, Red for asks
        self.setForeground(color)

class OrderBookWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("LeGoTradingEngine Orderbook Display")
        self.setGeometry(100, 100, 1000, 600)
        
        # Set dark theme
        self.setStyleSheet("""
            QMainWindow {
                background-color: #2c3e50;
            }
            QTableWidget {
                background-color: #34495e;
                gridline-color: #7f8c8d;
                color: white;
                border: none;
                border-radius: 5px;
            }
            QTableWidget::item {
                padding: 5px;
            }
            QTableWidget::item:selected {
                background-color: #2980b9;
            }
            QHeaderView::section {
                background-color: #2c3e50;
                color: white;
                padding: 5px;
                border: none;
            }
            QLabel {
                color: white;
                font-size: 14px;
                font-weight: bold;
                padding: 10px;
            }
        """)

        # Create main widget and horizontal layout
        main_widget = QWidget()
        self.setCentralWidget(main_widget)
        main_layout = QHBoxLayout(main_widget)

        # Create left widget for orderbook
        orderbook_widget = QWidget()
        orderbook_layout = QVBoxLayout(orderbook_widget)
        orderbook_layout.setSpacing(10)
        orderbook_layout.setContentsMargins(20, 20, 20, 20)

        # Add labels with updated styling
        asks_label = QLabel("ASKS (SELL ORDERS)")
        bids_label = QLabel("BIDS (BUY ORDERS)")
        for label in [asks_label, bids_label]:
            label.setAlignment(Qt.AlignmentFlag.AlignCenter)

        # Create tables with updated styling
        self.asks_table = QTableWidget(0, 2)
        self.bids_table = QTableWidget(0, 2)
        
        for table in [self.asks_table, self.bids_table]:
            table.setHorizontalHeaderLabels(["Price", "Quantity"])
            table.horizontalHeader().setStretchLastSection(True)
            table.verticalHeader().setVisible(False)
            table.setShowGrid(True)
            table.horizontalHeader().setDefaultAlignment(Qt.AlignmentFlag.AlignRight)
            table.setSelectionMode(QTableWidget.SelectionMode.NoSelection)
            table.setFocusPolicy(Qt.FocusPolicy.NoFocus)

        # Add orderbook widgets to left layout
        orderbook_layout.addWidget(asks_label)
        orderbook_layout.addWidget(self.asks_table)
        orderbook_layout.addSpacing(20)
        orderbook_layout.addWidget(bids_label)
        orderbook_layout.addWidget(self.bids_table)

        # Create right widget for metrics
        metrics_widget = QWidget()
        metrics_layout = QVBoxLayout(metrics_widget)
        metrics_layout.setSpacing(10)
        metrics_layout.setContentsMargins(20, 20, 20, 20)

        # Create labels for metrics
        metrics_label = QLabel("MARKET METRICS")
        metrics_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.last_price = QLabel("Last Price: --")
        self.best_bid = QLabel("Best Bid: --")
        self.best_ask = QLabel("Best Ask: --")
        self.spread = QLabel("Spread: --")

        # Add metrics to right layout
        metrics_layout.addWidget(metrics_label)
        for label in [self.last_price, self.best_bid, self.best_ask, self.spread]:
            label.setAlignment(Qt.AlignmentFlag.AlignLeft)
            metrics_layout.addWidget(label)
        
        # Add both widgets to main layout
        main_layout.addWidget(orderbook_widget, 2)
        main_layout.addWidget(metrics_widget, 1)

        # Setup ZMQ worker
        self.worker = OrderBookWorker()
        self.worker.update_signal.connect(self.update_orderbook)
        self.worker.start()

    def update_orderbook(self, state: OrderBookState):
        # Update asks (reverse order to show highest ask first)
        self.asks_table.setRowCount(len(state.asks))
        for i, ask in enumerate(reversed(state.asks)):
            self.asks_table.setItem(i, 0, StyledTableItem(f"{ask.price:.2f}", False))
            self.asks_table.setItem(i, 1, StyledTableItem(f"{ask.quantity:.2f}", False))

        # Update bids
        self.bids_table.setRowCount(len(state.bids))
        for i, bid in enumerate(state.bids):
            self.bids_table.setItem(i, 0, StyledTableItem(f"{bid.price:.2f}", True))
            self.bids_table.setItem(i, 1, StyledTableItem(f"{bid.quantity:.2f}", True))

        # Adjust columns to content
        self.asks_table.resizeColumnsToContents()
        self.bids_table.resizeColumnsToContents()

        # Update other metrics
        self.last_price.setText(f"Last Executed Price: {state.lastExecutedPrice}")
        self.best_bid.setText(f"Best Bid: {state.bestBid}")
        self.best_ask.setText(f"Best Ask: {state.bestAsk}")
        self.spread.setText(f"Spread: {state.spread}")

if __name__ == '__main__':
    app = QApplication(sys.argv)
    window = OrderBookWindow()
    window.show()
    sys.exit(app.exec())