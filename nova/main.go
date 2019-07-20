// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// geth is the official command-line client for Ethereum.
package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"path"
	godebug "runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/gosigar"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"sync"

	"github.com/manifoldco/promptui"
	demo "github.com/novaprotocolio/orderbook/common"
	"github.com/novaprotocolio/orderbook/orderbook"
	"github.com/novaprotocolio/orderbook/protocol"
	"github.com/novaprotocolio/orderbook/terminal"
)

// command line arguments
var (
	thisNode *node.Node
	// struct{} is the smallest data type available in Go, since it contains literally nothing
	quitC = make(chan struct{}, 1)

	// get the incoming message
	orderbookService protocol.OrderbookService	
	debugOrderbook = false
	OrderbookEnableDebugFlag = cli.BoolFlag{
		Name:  "obdebug",
		Usage: "Debug orderbook protocol",
	}
	prompt          *promptui.Select
	commands        []terminal.Command
	orderbookEngine *orderbook.Engine
)

const (
	clientIdentifier = "geth" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the go-ethereum command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.NoUSBFlag,
		utils.DashboardEnabledFlag,
		utils.DashboardAddrFlag,
		utils.DashboardPortFlag,
		utils.DashboardRefreshFlag,
		utils.EthashCacheDirFlag,
		utils.EthashCachesInMemoryFlag,
		utils.EthashCachesOnDiskFlag,
		utils.EthashDatasetDirFlag,
		utils.EthashDatasetsInMemoryFlag,
		utils.EthashDatasetsOnDiskFlag,
		utils.TxPoolLocalsFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		utils.LightServFlag,
		utils.LightPeersFlag,
		utils.LightKDFFlag,
		utils.WhitelistFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheTrieFlag,
		utils.CacheGCFlag,
		utils.TrieCacheGenFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.MiningEnabledFlag,
		utils.MinerThreadsFlag,
		utils.MinerLegacyThreadsFlag,
		utils.MinerNotifyFlag,
		utils.MinerGasTargetFlag,
		utils.MinerLegacyGasTargetFlag,
		utils.MinerGasLimitFlag,
		utils.MinerGasPriceFlag,
		utils.MinerLegacyGasPriceFlag,
		utils.MinerEtherbaseFlag,
		utils.MinerLegacyEtherbaseFlag,
		utils.MinerExtraDataFlag,
		utils.MinerLegacyExtraDataFlag,
		utils.MinerRecommitIntervalFlag,
		utils.MinerNoVerfiyFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.GoerliFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.ConstantinopleOverrideFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.EthStatsURLFlag,
		utils.MetricsEnabledFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		utils.GpoBlocksFlag,
		utils.GpoPercentileFlag,
		utils.EWASMInterpreterFlag,
		utils.EVMInterpreterFlag,
		OrderbookEnableDebugFlag,
		configFileFlag,
		utils.IstanbulRequestTimeoutFlag,
		utils.IstanbulBlockPeriodFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
		utils.RPCGlobalGasCap,
	}

	whisperFlags = []cli.Flag{
		utils.WhisperEnabledFlag,
		utils.WhisperMaxMessageSizeFlag,
		utils.WhisperMinPOWFlag,
		utils.WhisperRestrictConnectionBetweenLightClientsFlag,
	}

	metricsFlags = []cli.Flag{
		utils.MetricsEnableInfluxDBFlag,
		utils.MetricsInfluxDBEndpointFlag,
		utils.MetricsInfluxDBDatabaseFlag,
		utils.MetricsInfluxDBUsernameFlag,
		utils.MetricsInfluxDBPasswordFlag,
		utils.MetricsInfluxDBTagsFlag,
	}
)

func init() {
	// Initialize the CLI app and start Geth
	app.Action = geth
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2018 The go-ethereum Authors"
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand,
		importCommand,
		exportCommand,
		importPreimagesCommand,
		exportPreimagesCommand,
		copydbCommand,
		removedbCommand,
		dumpCommand,
		// See monitorcmd.go:
		monitorCommand,
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See misccmd.go:
		makecacheCommand,
		makedagCommand,
		versionCommand,
		bugCommand,
		licenseCommand,
		// See config.go
		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, whisperFlags...)
	app.Flags = append(app.Flags, metricsFlags...)

	app.Before = func(ctx *cli.Context) error {
		logdir := ""
		if ctx.GlobalBool(utils.DashboardEnabledFlag.Name) {
			logdir = (&node.Config{DataDir: utils.MakeDataDir(ctx)}).ResolvePath("logs")
		}
		if err := debug.Setup(ctx, logdir); err != nil {
			return err
		}
		// Cap the cache allowance and tune the garbage collector
		var mem gosigar.Mem
		if err := mem.Get(); err == nil {
			allowance := int(mem.Total / 1024 / 1024 / 3)
			if cache := ctx.GlobalInt(utils.CacheFlag.Name); cache > allowance {
				log.Warn("Sanitizing cache to Go's GC limits", "provided", cache, "updated", allowance)
				ctx.GlobalSet(utils.CacheFlag.Name, strconv.Itoa(allowance))
			}
		}
		// Ensure Go's GC ignores the database cache for trigger percentage
		cache := ctx.GlobalInt(utils.CacheFlag.Name)
		gogc := math.Max(20, math.Min(100, 100/(float64(cache)/1024)))

		log.Debug("Sanitizing Go's GC trigger", "percent", int(gogc))
		godebug.SetGCPercent(int(gogc))

		// Start metrics export if enabled
		utils.SetupMetrics(ctx)

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		// debug orderbook
		if ctx.GlobalBool(OrderbookEnableDebugFlag.Name) {
			debugOrderbook = true
		}

		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// geth is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func geth(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}
	node := makeFullNode(ctx)

	thisNode = node
	initPrompt()
	initOrderbook()

	startNode(ctx, node)

	startOrderbook()

	node.Wait()

	// commit changes to orderbook
	orderbookEngine.Commit()

	return nil
}

/** orderbook code **/

func shutdown() error {
	// return os.RemoveAll(thisNode.DataDir())
	return nil
}

func initPrompt() {

	cancelOrderArguments := []terminal.Argument{
		{Name: "order_id", Value: "1"},
		{Name: "pair_name", Value: "TOMO/WETH"},
		{Name: "side", Value: orderbook.Ask},
		{Name: "price", Value: "100"},
	}

	orderArguments := []terminal.Argument{
		{Name: "pair_name", Value: "TOMO/WETH"},
		{Name: "type", Value: "limit"},
		{Name: "side", Value: orderbook.Ask},
		{Name: "quantity", Value: "10"},
		{Name: "price", Value: "100", Hide: func(results map[string]string, thisArgument *terminal.Argument) bool {
			// ignore this argument when order type is market
			if results["type"] == "market" {
				return true
			}
			return false
		}},
		{Name: "trade_id", Value: "1"},
	}

	updateOrderArguments := append([]terminal.Argument{
		{Name: "order_id", Value: "1"},
	}, orderArguments...)

	// init prompt commands
	commands = []terminal.Command{
		{
			Name:        "addOrder",
			Arguments:   orderArguments,
			Description: "Add order",
		},
		{
			Name:        "updateOrder",
			Arguments:   updateOrderArguments,
			Description: "Update order, order_id must greater than 0",
		},
		{
			Name:        "cancelOrder",
			Arguments:   cancelOrderArguments,
			Description: "Cancel order, order_id must greater than 0",
		},
		{
			Name:        "nodeAddr",
			Description: "Get Node address",
		},
		{
			Name:        "quit",
			Description: "Quit the program",
		},
	}

	// cast type to sort
	// sort.Sort(terminal.CommandsByName(commands))

	prompt = terminal.NewPrompt("Your choice:", 6, commands)
}

func nodeAddr() string {
	return thisNode.Server().Self().String()
}

func processOrder(payload map[string]string) error {
	// add order at this current node first
	// get timestamp in milliseconds
	if payload["timestamp"] == "" {
		payload["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	}
	msg, err := protocol.NewOrderbookMsg(payload)
	if err == nil {
		// try to store into model, if success then process at local and broad cast
		trades, orderInBook := orderbookEngine.ProcessOrder(payload)
		demo.LogInfo("Orderbook result", "Trade", trades, "OrderInBook", orderInBook)

		// broad cast message
		orderbookService.OutC <- msg

	}

	return err
}

func cancelOrder(payload map[string]string) error {
	// add order at this current node first
	// get timestamp in milliseconds
	payload["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	msg, err := protocol.NewOrderbookMsg(payload)
	if err == nil {
		// try to store into model, if success then process at local and broad cast
		err := orderbookEngine.CancelOrder(payload)
		demo.LogInfo("Orderbook cancel result", "err", err, "msg", msg)

		// broad cast message
		orderbookService.OutC <- msg

	}

	return err
}

// simple ping and receive protocol
func initOrderbook() {

	var err error

	orderbookDir := path.Join(thisNode.DataDir(), "orderbook")

	// this should be from smart contract and update by multisig instead
	allowedPairs := map[string]*big.Int{
		"HOT-DAI": big.NewInt(10e9),
		"WETH-DAI": big.NewInt(10e9),
		"HOT-WETH": big.NewInt(10e9),
		"NOVA-WETH": big.NewInt(10e9),
	}
	orderbookEngine = orderbook.NewEngine(orderbookDir, allowedPairs)

	// register normal service, protocol is for p2p, service is for rpc calls
	service := protocol.NewService(quitC, orderbookEngine)
	err = thisNode.Register(service)

	if err != nil {
		demo.LogCrit("Register orderbook service in servicenode failed", "err", err)
	}

	if err != nil {
		demo.LogCrit(err.Error())
	}

}

func startOrderbook() error {

	// process command
	fmt.Println("---------------Welcome to Orderbook testing---------------------")

	// debug orderbook
	if debugOrderbook {	
		var endWaiter sync.WaitGroup
		endWaiter.Add(1)

		// start serving
		go func() {

			for {
				// loop command
				selected, _, err := prompt.Run()

				// unknow error, should retry
				if err != nil {
					demo.LogInfo("Prompt failed %v\n", err)
					continue
				}

				// get selected command and run it
				command := commands[selected]
				if command.Name == "quit" {
					demo.LogInfo("Server quiting...")					
					endWaiter.Done()
					thisNode.Stop()
					quitC <- struct{}{}
					demo.LogInfo("-> Goodbye\n")
					return
				}
				results := command.Run()

				// process command
				switch command.Name {
				case "addOrder":
					demo.LogInfo("-> Add order", "payload", results)
					// put message on channel
					results["order_id"] = "0"
					go processOrder(results)
				case "updateOrder":
					demo.LogInfo("-> Update order", "payload", results)
					// put message on channel
					go processOrder(results)
				case "cancelOrder":
					demo.LogInfo("-> Cancel order", "payload", results)
					// put message on channel
					go cancelOrder(results)

				case "nodeAddr":
					demo.LogInfo(fmt.Sprintf("-> Node Address: %s\n", nodeAddr()))

				default:
					demo.LogInfo(fmt.Sprintf("-> Unknown command: %s\n", command.Name))
				}
			}

		}()

		// wait for command processing
		endWaiter.Wait()
	} else {

		demo.WaitForCtrlC()
	}

	// finally shutdown
	return shutdown()
}

/** end orderbook code **/

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	passwords := utils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ctx, ks, trimmed, i, passwords)
		}
	}
	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create a chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := ethclient.NewClient(rpcClient)

		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				derivationPath := accounts.DefaultBaseDerivationPath
				if event.Wallet.URL().Scheme == "ledger" {
					derivationPath = accounts.DefaultLedgerBaseDerivationPath
				}
				event.Wallet.SelfDerive(derivationPath, stateReader)

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
	// Start auxiliary services if enabled
	if ctx.GlobalBool(utils.MiningEnabledFlag.Name) || ctx.GlobalBool(utils.DeveloperFlag.Name) {
		// Mining only makes sense if a full Ethereum node is running
		if ctx.GlobalString(utils.SyncModeFlag.Name) == "light" {
			utils.Fatalf("Light clients do not support mining")
		}
		var ethereum *eth.Ethereum
		if err := stack.Service(&ethereum); err != nil {
			utils.Fatalf("Ethereum service not running: %v", err)
		}
		// Set the gas price to the limits from the CLI and start mining
		gasprice := utils.GlobalBig(ctx, utils.MinerLegacyGasPriceFlag.Name)
		if ctx.IsSet(utils.MinerGasPriceFlag.Name) {
			gasprice = utils.GlobalBig(ctx, utils.MinerGasPriceFlag.Name)
		}
		ethereum.TxPool().SetGasPrice(gasprice)

		threads := ctx.GlobalInt(utils.MinerLegacyThreadsFlag.Name)
		if ctx.GlobalIsSet(utils.MinerThreadsFlag.Name) {
			threads = ctx.GlobalInt(utils.MinerThreadsFlag.Name)
		}
		if err := ethereum.StartMining(threads); err != nil {
			utils.Fatalf("Failed to start mining: %v", err)
		}
	}
}
