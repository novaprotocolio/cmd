module github.com/ethereum/go-ethereum/cmd/nova

go 1.12

require (
	github.com/novaprotocolio/orderbook v0.0.0-00010101000000-000000000000 // indirect
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
	github.com/gizak/termui v2.3.0+incompatible // indirect
	github.com/maruel/panicparse v1.2.1 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/nsf/termbox-go v0.0.0-20190624072549-eeb6cd0a1762 // indirect
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
)

// for local test only
replace github.com/ethereum/go-ethereum => ../../novalex // v1.8.27

replace github.com/novaprotocolio/orderbook => ../../orderbook

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
