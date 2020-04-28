// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package intprotocol implements the Ethereum protocol.
package intprotocol

import (
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/accounts"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft"
	tendermintBackend "github.com/intfoundation/intchain/consensus/ipbft"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/bloombits"
	"github.com/intfoundation/intchain/core/datareduction"
	"github.com/intfoundation/intchain/core/rawdb"
	"github.com/intfoundation/intchain/core/types"
	"github.com/intfoundation/intchain/core/vm"
	"github.com/intfoundation/intchain/event"
	"github.com/intfoundation/intchain/intdb"
	"github.com/intfoundation/intchain/internal/ethapi"
	"github.com/intfoundation/intchain/intprotocol/downloader"
	"github.com/intfoundation/intchain/intprotocol/filters"
	"github.com/intfoundation/intchain/intprotocol/gasprice"
	"github.com/intfoundation/intchain/log"
	"github.com/intfoundation/intchain/miner"
	"github.com/intfoundation/intchain/node"
	"github.com/intfoundation/intchain/p2p"
	"github.com/intfoundation/intchain/params"
	"github.com/intfoundation/intchain/rlp"
	"github.com/intfoundation/intchain/rpc"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// IntChain implements the INT Chain full node service.
type IntChain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the intChain

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager

	// DB interfaces
	chainDb intdb.Database // Block chain database
	pruneDb intdb.Database // Prune data database

	eventMux       *event.TypeMux
	engine         consensus.IPBFT
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *EthApiBackend

	miner    *miner.Miner
	gasPrice *big.Int
	coinbase common.Address
	solcPath string

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
}

// New creates a new INT Chain object (including the
// initialisation of the common INT Chain object)
func New(ctx *node.ServiceContext, config *Config, cliCtx *cli.Context,
	cch core.CrossChainHelper, logger log.Logger, isTestnet bool) (*IntChain, error) {

	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := ctx.OpenDatabase("chaindata", config.DatabaseCache, config.DatabaseHandles, "eth/db/chaindata/")
	if err != nil {
		return nil, err
	}
	pruneDb, err := ctx.OpenDatabase("prunedata", config.DatabaseCache, config.DatabaseHandles, "intchain/db/prune/")
	if err != nil {
		return nil, err
	}

	isMainChain := params.IsMainChain(ctx.ChainId())

	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockWithDefault(chainDb, config.Genesis, isMainChain, isTestnet)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	chainConfig.ChainLogger = logger
	logger.Info("Initialised chain configuration", "config", chainConfig)

	eth := &IntChain{
		config:         config,
		chainDb:        chainDb,
		pruneDb:        pruneDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDb, cliCtx, cch),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.MinerGasPrice,
		coinbase:       config.Coinbase,
		solcPath:       config.SolcPath,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	//logger.Info("Initialising IntChain protocol", "versions", eth.engine.Protocol().Versions, "network", config.NetworkId, "dbversion", dbVer)
	logger.Info("Initialising IntChain protocol", "network", config.NetworkId, "dbversion", dbVer)

	if !config.SkipBcVersionCheck {
		if bcVersion != nil && *bcVersion > core.BlockChainVersion {
			return nil, fmt.Errorf("database version is v%d, Geth %s only supports v%d", *bcVersion, params.VersionWithMeta, core.BlockChainVersion)
		} else if bcVersion == nil || *bcVersion < core.BlockChainVersion {
			logger.Warn("Upgrade blockchain database version", "from", dbVer, "to", core.BlockChainVersion)
			rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
		}
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{
			TrieCleanLimit: config.TrieCleanCache,

			TrieDirtyLimit:    config.TrieDirtyCache,
			TrieDirtyDisabled: config.NoPruning,
			TrieTimeLimit:     config.TrieTimeout,
		}
	)
	//eth.engine = CreateConsensusEngine(ctx, config, chainConfig, chainDb, cliCtx, cch)

	eth.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, eth.chainConfig, eth.engine, vmConfig, cch)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		logger.Warn("Rewinding chain to upgrade configuration", "err", compat)
		eth.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	eth.bloomIndexer.Start(eth.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	eth.txPool = core.NewTxPool(config.TxPool, eth.chainConfig, eth.blockchain, cch)

	if eth.protocolManager, err = NewProtocolManager(eth.chainConfig, config.SyncMode, config.NetworkId, eth.eventMux, eth.txPool, eth.engine, eth.blockchain, chainDb, cch); err != nil {
		return nil, err
	}
	eth.miner = miner.New(eth, eth.chainConfig, eth.EventMux(), eth.engine, config.MinerGasFloor, config.MinerGasCeil, cch)
	eth.miner.SetExtra(makeExtraData(config.ExtraData))

	eth.ApiBackend = &EthApiBackend{eth, nil, nil, cch}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.MinerGasPrice
	}
	eth.ApiBackend.gpo = gasprice.NewOracle(eth.ApiBackend, gpoParams)

	return eth, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"intchain",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateConsensusEngine creates the required type of consensus engine instance for an IntChain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db intdb.Database,
	cliCtx *cli.Context, cch core.CrossChainHelper) consensus.IPBFT {
	// If Tendermint is requested, set it up
	if chainConfig.IPBFT.Epoch != 0 {
		config.IPBFT.Epoch = chainConfig.IPBFT.Epoch
	}
	config.IPBFT.ProposerPolicy = ipbft.ProposerPolicy(chainConfig.IPBFT.ProposerPolicy)
	return tendermintBackend.New(chainConfig, cliCtx, ctx.NodeKey(), cch)
}

// APIs returns the collection of RPC services the IntChain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *IntChain) APIs() []rpc.API {

	apis := ethapi.GetAPIs(s.ApiBackend, s.solcPath)
	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)
	// Append all the local APIs and return
	apis = append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "int",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "int",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "int",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "int",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
	return apis
}

func (s *IntChain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *IntChain) Coinbase() (eb common.Address, err error) {
	if ipbft, ok := s.engine.(consensus.IPBFT); ok {
		eb = ipbft.PrivateValidator()
		if eb != (common.Address{}) {
			return eb, nil
		} else {
			return eb, errors.New("private validator missing")
		}
	} else {
		s.lock.RLock()
		coinbase := s.coinbase
		s.lock.RUnlock()

		if coinbase != (common.Address{}) {
			return coinbase, nil
		}
		if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				coinbase := accounts[0].Address

				s.lock.Lock()
				s.coinbase = coinbase
				s.lock.Unlock()

				log.Info("Coinbase automatically configured", "address", coinbase)
				return coinbase, nil
			}
		}
	}
	return common.Address{}, fmt.Errorf("etherbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *IntChain) SetCoinbase(coinbase common.Address) {

	self.lock.Lock()
	self.coinbase = coinbase
	self.lock.Unlock()

	self.miner.SetCoinbase(coinbase)
}

func (s *IntChain) StartMining(local bool) error {
	var eb common.Address
	if ipbft, ok := s.engine.(consensus.IPBFT); ok {
		eb = ipbft.PrivateValidator()
		if (eb == common.Address{}) {
			log.Error("Cannot start mining without private validator")
			return errors.New("private validator missing")
		}
	} else {
		_, err := s.Coinbase()
		if err != nil {
			log.Error("Cannot start mining without etherbase", "err", err)
			return fmt.Errorf("etherbase missing: %v", err)
		}
	}

	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *IntChain) StopMining()         { s.miner.Stop() }
func (s *IntChain) IsMining() bool      { return s.miner.Mining() }
func (s *IntChain) Miner() *miner.Miner { return s.miner }

func (s *IntChain) ChainConfig() *params.ChainConfig   { return s.chainConfig }
func (s *IntChain) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *IntChain) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *IntChain) TxPool() *core.TxPool               { return s.txPool }
func (s *IntChain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *IntChain) Engine() consensus.IPBFT            { return s.engine }
func (s *IntChain) ChainDb() intdb.Database            { return s.chainDb }
func (s *IntChain) IsListening() bool                  { return true } // Always listening
func (s *IntChain) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *IntChain) NetVersion() uint64                 { return s.networkId }
func (s *IntChain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *IntChain) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// IntChain protocol implementation.
func (s *IntChain) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers

	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)

	// Start the Auto Mining Loop
	go s.loopForMiningEvent()

	// Start the Data Reduction
	if s.config.PruneStateData && s.chainConfig.IntChainId == "child_0" {
		go s.StartScanAndPrune(0)
	}

	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// IntChain protocol.
func (s *IntChain) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.miner.Stop()
	s.engine.Close()
	s.miner.Close()
	s.eventMux.Stop()

	s.chainDb.Close()
	s.pruneDb.Close()
	close(s.shutdownChan)

	return nil
}

func (s *IntChain) loopForMiningEvent() {
	// Start/Stop mining Feed
	startMiningCh := make(chan core.StartMiningEvent, 1)
	startMiningSub := s.blockchain.SubscribeStartMiningEvent(startMiningCh)

	stopMiningCh := make(chan core.StopMiningEvent, 1)
	stopMiningSub := s.blockchain.SubscribeStopMiningEvent(stopMiningCh)

	defer startMiningSub.Unsubscribe()
	defer stopMiningSub.Unsubscribe()

	for {
		select {
		case <-startMiningCh:
			if !s.IsMining() {
				s.lock.RLock()
				price := s.gasPrice
				s.lock.RUnlock()
				s.txPool.SetGasPrice(price)
				s.chainConfig.ChainLogger.Info("IPBFT Consensus Engine will be start shortly")
				s.engine.(consensus.IPBFT).ForceStart()
				s.StartMining(true)
			} else {
				s.chainConfig.ChainLogger.Info("IPBFT Consensus Engine already started")
			}
		case <-stopMiningCh:
			if s.IsMining() {
				s.chainConfig.ChainLogger.Info("IPBFT Consensus Engine will be stop shortly")
				s.StopMining()
			} else {
				s.chainConfig.ChainLogger.Info("IPBFT Consensus Engine already stopped")
			}
		case <-startMiningSub.Err():
			return
		case <-stopMiningSub.Err():
			return
		}
	}
}

func (s *IntChain) StartScanAndPrune(blockNumber uint64) {

	if datareduction.StartPruning() {
		log.Info("Data Reduction - Start")
	} else {
		log.Info("Data Reduction - Pruning is already running")
		return
	}

	latestBlockNumber := s.blockchain.CurrentHeader().Number.Uint64()
	if blockNumber == 0 || blockNumber >= latestBlockNumber {
		blockNumber = latestBlockNumber
		log.Infof("Data Reduction - Last block number %v", blockNumber)
	} else {
		log.Infof("Data Reduction - User defined Last block number %v", blockNumber)
	}

	ps := rawdb.ReadHeadScanNumber(s.pruneDb)
	var scanNumber uint64
	if ps != nil {
		scanNumber = *ps
	}

	pp := rawdb.ReadHeadPruneNumber(s.pruneDb)
	var pruneNumber uint64
	if pp != nil {
		pruneNumber = *pp
	}
	log.Infof("Data Reduction - Last scan number %v, prune number %v", scanNumber, pruneNumber)

	pruneProcessor := datareduction.NewPruneProcessor(s.chainDb, s.pruneDb, s.blockchain, s.config.PruneBlockData)

	lastScanNumber, lastPruneNumber := pruneProcessor.Process(blockNumber, scanNumber, pruneNumber)
	log.Infof("Data Reduction - After prune, last number scan %v, prune number %v", lastScanNumber, lastPruneNumber)
	if s.config.PruneBlockData {
		for i := uint64(1); i < lastPruneNumber; i++ {
			rawdb.DeleteBody(s.chainDb, rawdb.ReadCanonicalHash(s.chainDb, i), i)
		}
		log.Infof("deleted block from 1 to %v", lastPruneNumber)
	}
	log.Info("Data Reduction - Completed")

	datareduction.StopPruning()
}
