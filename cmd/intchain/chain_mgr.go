package main

import (
	"github.com/intfoundation/go-crypto"
	dbm "github.com/intfoundation/go-db"
	"github.com/intfoundation/intchain/accounts"
	"github.com/intfoundation/intchain/cmd/utils"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft/epoch"
	"github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/rawdb"
	"github.com/intfoundation/intchain/intclient"
	"github.com/intfoundation/intchain/intp2p"
	"github.com/intfoundation/intchain/intprotocol"
	"github.com/intfoundation/intchain/intrpc"
	"github.com/intfoundation/intchain/log"
	"github.com/intfoundation/intchain/node"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"sync"
)

type ChainManager struct {
	ctx *cli.Context

	mainChain     *Chain
	mainQuit      <-chan struct{}
	mainStartDone chan struct{}

	createChildChainLock sync.Mutex
	childChains          map[string]*Chain
	childQuits           map[string]<-chan struct{}

	stop chan struct{} // Channel wait for INTCHAIN stop

	server *intp2p.IntChainP2PServer
	cch    *CrossChainHelper
}

var chainMgr *ChainManager
var once sync.Once

func GetCMInstance(ctx *cli.Context) *ChainManager {

	once.Do(func() {
		chainMgr = &ChainManager{ctx: ctx}
		chainMgr.stop = make(chan struct{})
		chainMgr.childChains = make(map[string]*Chain)
		chainMgr.childQuits = make(map[string]<-chan struct{})
		chainMgr.cch = &CrossChainHelper{}
	})
	return chainMgr
}

func (cm *ChainManager) GetNodeID() string {
	return cm.server.Server().NodeInfo().ID
}

func (cm *ChainManager) InitP2P() {
	cm.server = intp2p.NewP2PServer(cm.ctx)
}

func (cm *ChainManager) LoadMainChain() error {
	// Load Main Chain
	chainId := MainChain
	if cm.ctx.GlobalBool(utils.TestnetFlag.Name) {
		chainId = TestnetChain
	}
	cm.mainChain = LoadMainChain(cm.ctx, chainId)
	if cm.mainChain == nil {
		return errors.New("Load main chain failed")
	}

	return nil
}

func (cm *ChainManager) LoadChains(childIds []string) error {

	childChainIds := core.GetChildChainIds(cm.cch.chainInfoDB)
	log.Infof("Before load child chains, child chain IDs are %v, len is %d", childChainIds, len(childChainIds))

	readyToLoadChains := make(map[string]bool) // Key: Child Chain ID, Value: Enable Mining (deprecated)

	// Check we are belong to the validator of Child Chain in DB first (Mining Mode)
	for _, chainId := range childChainIds {
		// Check Current Validator is Child Chain Validator
		ci := core.GetChainInfo(cm.cch.chainInfoDB, chainId)
		// Check if we are in this child chain
		if ci.Epoch != nil && cm.checkCoinbaseInChildChain(ci.Epoch) {
			readyToLoadChains[chainId] = true
		}
	}

	// Check request from Child Chain
	for _, requestId := range childIds {
		if requestId == "" {
			// Ignore the Empty ID
			continue
		}

		if _, present := readyToLoadChains[requestId]; present {
			// Already loaded, ignore
			continue
		} else {
			// Launch in non-mining mode, including both correct and wrong chain id
			// Wrong chain id will be ignore after loading failed
			readyToLoadChains[requestId] = false
		}
	}

	log.Infof("Number of child chain to be loaded :%v", len(readyToLoadChains))
	log.Infof("Start to load child chain: %v", readyToLoadChains)

	for chainId := range readyToLoadChains {
		chain := LoadChildChain(cm.ctx, chainId)
		if chain == nil {
			log.Errorf("Load child chain: %s Failed.", chainId)
			continue
		}

		cm.childChains[chainId] = chain
		log.Infof("Load child chain: %s Success!", chainId)
	}
	return nil
}

func (cm *ChainManager) InitCrossChainHelper() {
	cm.cch.chainInfoDB = dbm.NewDB("chaininfo",
		cm.mainChain.Config.GetString("db_backend"),
		cm.ctx.GlobalString(utils.DataDirFlag.Name))
	cm.cch.localTX3CacheDB, _ = rawdb.NewLevelDBDatabase(path.Join(cm.ctx.GlobalString(utils.DataDirFlag.Name), "tx3cache"), 0, 0, "intchain/db/tx3/")

	chainId := MainChain
	if cm.ctx.GlobalBool(utils.TestnetFlag.Name) {
		chainId = TestnetChain
	}
	cm.cch.mainChainId = chainId

	if cm.ctx.GlobalBool(utils.RPCEnabledFlag.Name) {
		host := "127.0.0.1" //cm.ctx.GlobalString(utils.RPCListenAddrFlag.Name)
		port := cm.ctx.GlobalInt(utils.RPCPortFlag.Name)
		url := net.JoinHostPort(host, strconv.Itoa(port))
		url = "http://" + url + "/" + chainId
		client, err := intclient.Dial(url)
		if err != nil {
			log.Errorf("can't connect to %s, err: %v, exit", url, err)
			os.Exit(0)
		}
		cm.cch.client = client
	}
}

func (cm *ChainManager) StartP2PServer() error {
	srv := cm.server.Server()
	// Append Main Chain Protocols
	srv.Protocols = append(srv.Protocols, cm.mainChain.IntNode.GatherProtocols()...)
	// Append Child Chain Protocols
	//for _, chain := range cm.childChains {
	//	srv.Protocols = append(srv.Protocols, chain.EthNode.GatherProtocols()...)
	//}
	// Start the server
	return srv.Start()
}

func (cm *ChainManager) StartMainChain() error {
	// Start the Main Chain
	cm.mainStartDone = make(chan struct{})

	cm.mainChain.IntNode.SetP2PServer(cm.server.Server())

	if address, ok := cm.getNodeValidator(cm.mainChain.IntNode); ok {
		cm.server.AddLocalValidator(cm.mainChain.Id, address)
	}

	err := StartChain(cm.ctx, cm.mainChain, cm.mainStartDone)

	// Wait for Main Chain Start Complete
	<-cm.mainStartDone
	cm.mainQuit = cm.mainChain.IntNode.StopChan()

	return err
}

func (cm *ChainManager) StartChains() error {

	for _, chain := range cm.childChains {
		// Start each Chain
		srv := cm.server.Server()
		childProtocols := chain.IntNode.GatherProtocols()
		// Add Child Protocols to P2P Server Protocols
		srv.Protocols = append(srv.Protocols, childProtocols...)
		// Add Child Protocols to P2P Server Caps
		srv.AddChildProtocolCaps(childProtocols)

		chain.IntNode.SetP2PServer(srv)

		if address, ok := cm.getNodeValidator(chain.IntNode); ok {
			cm.server.AddLocalValidator(chain.Id, address)
		}

		startDone := make(chan struct{})
		StartChain(cm.ctx, chain, startDone)
		<-startDone

		cm.childQuits[chain.Id] = chain.IntNode.StopChan()

		// Tell other peers that we have added into a new child chain
		cm.server.BroadcastNewChildChainMsg(chain.Id)
	}

	return nil
}

func (cm *ChainManager) StartRPC() error {

	// Start IntChain RPC
	err := intrpc.StartRPC(cm.ctx)
	if err != nil {
		return err
	} else {
		if intrpc.IsHTTPRunning() {
			if h, err := cm.mainChain.IntNode.GetHTTPHandler(); err == nil {
				intrpc.HookupHTTP(cm.mainChain.Id, h)
			} else {
				log.Errorf("Load Main Chain RPC HTTP handler failed: %v", err)
			}
			for _, chain := range cm.childChains {
				if h, err := chain.IntNode.GetHTTPHandler(); err == nil {
					intrpc.HookupHTTP(chain.Id, h)
				} else {
					log.Errorf("Load Child Chain RPC HTTP handler failed: %v", err)
				}
			}
		}

		if intrpc.IsWSRunning() {
			if h, err := cm.mainChain.IntNode.GetWSHandler(); err == nil {
				intrpc.HookupWS(cm.mainChain.Id, h)
			} else {
				log.Errorf("Load Main Chain RPC WS handler failed: %v", err)
			}
			for _, chain := range cm.childChains {
				if h, err := chain.IntNode.GetWSHandler(); err == nil {
					intrpc.HookupWS(chain.Id, h)
				} else {
					log.Errorf("Load Child Chain RPC WS handler failed: %v", err)
				}
			}
		}
	}

	return nil
}

func (cm *ChainManager) StartInspectEvent() {

	createChildChainCh := make(chan core.CreateChildChainEvent, 10)
	createChildChainSub := MustGetIntChainFromNode(cm.mainChain.IntNode).BlockChain().SubscribeCreateChildChainEvent(createChildChainCh)

	go func() {
		defer createChildChainSub.Unsubscribe()

		for {
			select {
			case event := <-createChildChainCh:
				log.Infof("CreateChildChainEvent received: %v", event)

				go func() {
					cm.createChildChainLock.Lock()
					defer cm.createChildChainLock.Unlock()

					cm.LoadChildChainInRT(event.ChainId)
				}()
			case <-createChildChainSub.Err():
				return
			}
		}
	}()
}

func (cm *ChainManager) LoadChildChainInRT(chainId string) {

	// Load Child Chain data from pending data
	cci := core.GetPendingChildChainData(cm.cch.chainInfoDB, chainId)
	if cci == nil {
		log.Errorf("child chain: %s does not exist, can't load", chainId)
		return
	}

	validators := make([]types.GenesisValidator, 0, len(cci.JoinedValidators))

	validator := false

	var ethereum *intprotocol.IntChain
	cm.mainChain.IntNode.Service(&ethereum)

	var localEtherbase common.Address
	if ipbft, ok := ethereum.Engine().(consensus.IPBFT); ok {
		localEtherbase = ipbft.PrivateValidator()
	}

	for _, v := range cci.JoinedValidators {
		if v.Address == localEtherbase {
			validator = true
		}

		// dereference the PubKey
		if pubkey, ok := v.PubKey.(*crypto.BLSPubKey); ok {
			v.PubKey = *pubkey
		}

		// append the Validator
		validators = append(validators, types.GenesisValidator{
			EthAccount: v.Address,
			PubKey:     v.PubKey,
			Amount:     v.DepositAmount,
		})
	}

	// Write down the genesis into chain info db when exit the routine
	defer writeGenesisIntoChainInfoDB(cm.cch.chainInfoDB, chainId, validators)

	if !validator {
		log.Warnf("You are not in the validators of child chain %v, no need to start the child chain", chainId)
		// Update Child Chain to formal
		cm.formalizeChildChain(chainId, *cci, nil)
		return
	}

	// if child chain already loaded, just return (For catch-up case)
	if _, ok := cm.childChains[chainId]; ok {
		log.Infof("Child Chain [%v] has been already loaded.", chainId)
		return
	}

	// Load the KeyStore file from MainChain (Optional)
	var keyJson []byte
	wallet, walletErr := cm.mainChain.IntNode.AccountManager().Find(accounts.Account{Address: localEtherbase})
	if walletErr == nil {
		var readKeyErr error
		keyJson, readKeyErr = ioutil.ReadFile(wallet.URL().Path)
		if readKeyErr != nil {
			log.Errorf("Failed to Read the KeyStore %v, Error: %v", localEtherbase, readKeyErr)
		}
	}

	// child chain uses the same validator with the main chain.
	privValidatorFile := cm.mainChain.Config.GetString("priv_validator_file")
	self := types.LoadPrivValidator(privValidatorFile)

	err := CreateChildChain(cm.ctx, chainId, *self, keyJson, validators)
	if err != nil {
		log.Errorf("Create Child Chain %v failed! %v", chainId, err)
		return
	}

	chain := LoadChildChain(cm.ctx, chainId)
	if chain == nil {
		log.Errorf("Child Chain %v load failed!", chainId)
		return
	}

	//StartChildChain to attach intp2p and intrpc
	//TODO Hookup new Created Child Chain to P2P server
	srv := cm.server.Server()
	childProtocols := chain.IntNode.GatherProtocols()
	// Add Child Protocols to P2P Server Protocols
	srv.Protocols = append(srv.Protocols, childProtocols...)
	// Add Child Protocols to P2P Server Caps
	srv.AddChildProtocolCaps(childProtocols)

	chain.IntNode.SetP2PServer(srv)

	if address, ok := cm.getNodeValidator(chain.IntNode); ok {
		srv.AddLocalValidator(chain.Id, address)
	}

	// Start the new Child Chain, and it will start child chain reactors as well
	startDone := make(chan struct{})
	err = StartChain(cm.ctx, chain, startDone)
	<-startDone
	if err != nil {
		return
	}

	cm.childQuits[chain.Id] = chain.IntNode.StopChan()

	var childEthereum *intprotocol.IntChain
	chain.IntNode.Service(&childEthereum)
	firstEpoch := childEthereum.Engine().(consensus.IPBFT).GetEpoch()
	// Child Chain start success, then delete the pending data in chain info db
	cm.formalizeChildChain(chainId, *cci, firstEpoch)

	// Add Child Chain Id into Chain Manager
	cm.childChains[chainId] = chain

	//TODO Broadcast Child ID to all Main Chain peers
	go cm.server.BroadcastNewChildChainMsg(chainId)

	//hookup intrpc
	if intrpc.IsHTTPRunning() {
		if h, err := chain.IntNode.GetHTTPHandler(); err == nil {
			intrpc.HookupHTTP(chain.Id, h)
		} else {
			log.Errorf("Unable Hook up Child Chain (%v) RPC HTTP Handler: %v", chainId, err)
		}
	}
	if intrpc.IsWSRunning() {
		if h, err := chain.IntNode.GetWSHandler(); err == nil {
			intrpc.HookupWS(chain.Id, h)
		} else {
			log.Errorf("Unable Hook up Child Chain (%v) RPC WS Handler: %v", chainId, err)
		}
	}

}

func (cm *ChainManager) formalizeChildChain(chainId string, cci core.CoreChainInfo, ep *epoch.Epoch) {
	// Child Chain start success, then delete the pending data in chain info db
	core.DeletePendingChildChainData(cm.cch.chainInfoDB, chainId)
	// Convert the Chain Info from Pending to Formal
	core.SaveChainInfo(cm.cch.chainInfoDB, &core.ChainInfo{CoreChainInfo: cci, Epoch: ep})
}

func (cm *ChainManager) checkCoinbaseInChildChain(childEpoch *epoch.Epoch) bool {
	var ethereum *intprotocol.IntChain
	cm.mainChain.IntNode.Service(&ethereum)

	var localEtherbase common.Address
	if ipbft, ok := ethereum.Engine().(consensus.IPBFT); ok {
		localEtherbase = ipbft.PrivateValidator()
	}

	return childEpoch.Validators.HasAddress(localEtherbase[:])
}

func (cm *ChainManager) StopChain() {
	go func() {
		mainChainError := cm.mainChain.IntNode.Close()
		if mainChainError != nil {
			log.Error("Error when closing main chain", "err", mainChainError)
		} else {
			log.Info("Main Chain Closed")
		}
	}()
	for _, child := range cm.childChains {
		go func() {
			childChainError := child.IntNode.Close()
			if childChainError != nil {
				log.Error("Error when closing child chain", "child id", child.Id, "err", childChainError)
			}
		}()
	}
}

func (cm *ChainManager) WaitChainsStop() {
	<-cm.mainQuit
	for _, quit := range cm.childQuits {
		<-quit
	}
}

func (cm *ChainManager) Stop() {
	intrpc.StopRPC()
	cm.server.Stop()
	cm.cch.localTX3CacheDB.Close()
	cm.cch.chainInfoDB.Close()

	// Release the main routine
	close(cm.stop)
}

func (cm *ChainManager) Wait() {
	<-cm.stop
}

func (cm *ChainManager) getNodeValidator(intNode *node.Node) (common.Address, bool) {

	var intchain *intprotocol.IntChain
	intNode.Service(&intchain)

	var coinbase common.Address
	tdm := intchain.Engine()
	epoch := tdm.GetEpoch()
	coinbase = tdm.PrivateValidator()
	log.Debugf("getNodeValidator() coinbase is :%v", coinbase)
	return coinbase, epoch.Validators.HasAddress(coinbase[:])
}

func writeGenesisIntoChainInfoDB(db dbm.DB, childChainId string, validators []types.GenesisValidator) {
	ethByte, _ := generateETHGenesis(childChainId, validators)
	tdmByte, _ := generateTDMGenesis(childChainId, validators)
	core.SaveChainGenesis(db, childChainId, ethByte, tdmByte)
}
