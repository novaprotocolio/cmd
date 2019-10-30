module github.com/ethereum/go-ethereum/cmd/nova

go 1.13

require (
	github.com/elastic/gosigar v0.10.5
	github.com/ethereum/go-ethereum v1.9.6
	github.com/gizak/termui v2.2.1-0.20170117222342-991cd3d38091+incompatible
	github.com/manifoldco/promptui v0.3.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/novaprotocolio/orderbook v0.0.0-00010101000000-000000000000
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/syndtr/goleveldb v1.0.0
	gopkg.in/urfave/cli.v1 v1.20.0
)

// for local test only
replace github.com/ethereum/go-ethereum => ../../novalex // v1.8.27

replace github.com/novaprotocolio/orderbook => ../../orderbook

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
