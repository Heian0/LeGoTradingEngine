package orderbook

import "strconv"

type Symbol struct {
	symbolId uint64
	ticker   string
}

func NewSymbol(_symbolId uint64, _ticker string) Symbol {
	return Symbol{
		symbolId: _symbolId,
		ticker:   _ticker,
	}
}

func (symbol *Symbol) String() string {
	symbolString := strconv.FormatUint(symbol.symbolId, 10) + " " + symbol.ticker + "\n"
	return symbolString
}
