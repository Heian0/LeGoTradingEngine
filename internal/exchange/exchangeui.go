package exchange

import (
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type ExchangeUI struct {
	ui       *app.Window
	exchange *Exchange
	th       *material.Theme
}

func (exchange *Exchange) InitializeExchangeUI() {
	w := new(app.Window)
	w.Option(app.Title("NYSE"))
	w.Option(app.Size(unit.Dp(800), unit.Dp(600)))

	exgUI := &ExchangeUI{
		ui:       w,
		th:       material.NewTheme(),
		exchange: exchange,
	}
	exchange.ui = exgUI

	go exgUI.Run() // Run UI in separate goroutine
}

// run handles the main UI loop
func (ew *ExchangeUI) Run() {
	var ops op.Ops

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		// Handle events
		e := ew.ui.Event()
		if e != nil {
			switch e := e.(type) {
			case app.DestroyEvent:
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				ew.layout(gtx)
				e.Frame(gtx.Ops)
			}
		}

		// Check ticker
		select {
		case <-ticker.C:
			ew.ui.Invalidate()
		default:
		}
	}
}

// layout defines the UI layout
func (ew *ExchangeUI) layout(gtx layout.Context) layout.Dimensions {
	// Lock the exchange while reading
	ew.exchange.RLock()
	// Use the data directly from the exchange pointer
	data := ew.exchange.String()
	ew.exchange.RUnlock()

	// Create a simple label with the exchange data
	label := material.H3(ew.th, data)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, label.Layout)
		}),
	)
}
