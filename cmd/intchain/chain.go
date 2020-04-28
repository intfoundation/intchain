package main

import (
	cfg "github.com/intfoundation/go-config"
	"github.com/intfoundation/intchain/accounts/keystore"
	"github.com/intfoundation/intchain/cmd/utils"
	tdmTypes "github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/log"
	intnode "github.com/intfoundation/intchain/node"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
)

const (
	// Client identifier to advertise over the network
	MainChain    = "intchain"
	TestnetChain = "testnet"
)

type Chain struct {
	Id      string
	Config  cfg.Config
	IntNode *intnode.Node
}

func LoadMainChain(ctx *cli.Context, chainId string) *Chain {

	chain := &Chain{Id: chainId}
	config := utils.GetTendermintConfig(chainId, ctx)
	chain.Config = config

	log.Info("Make full node")
	stack := makeFullNode(ctx, GetCMInstance(ctx).cch)
	chain.IntNode = stack

	return chain
}

func LoadChildChain(ctx *cli.Context, chainId string) *Chain {

	log.Infof("now load child: %s", chainId)

	//chainDir := ChainDir(ctx, chainId)
	//empty, err := cmn.IsDirEmpty(chainDir)
	//log.Infof("chainDir is : %s, empty is %v", chainDir, empty)
	//if empty || err != nil {
	//	log.Errorf("directory %s not exist or with error %v", chainDir, err)
	//	return nil
	//}
	chain := &Chain{Id: chainId}
	config := utils.GetTendermintConfig(chainId, ctx)
	chain.Config = config

	log.Infof("chainId: %s, makeFullNode", chainId)
	cch := GetCMInstance(ctx).cch
	stack := makeFullNode(ctx, cch)
	if stack == nil {
		return nil
	} else {
		chain.IntNode = stack
		return chain
	}
}

func StartChain(ctx *cli.Context, chain *Chain, startDone chan<- struct{}) error {

	log.Infof("Start Chain: %s", chain.Id)
	go func() {
		log.Info("StartChain()->utils.StartNode(stack)")
		utils.StartNodeEx(ctx, chain.IntNode)

		if startDone != nil {
			startDone <- struct{}{}
		}
	}()

	return nil
}

func CreateChildChain(ctx *cli.Context, chainId string, validator tdmTypes.PrivValidator, keyJson []byte, validators []tdmTypes.GenesisValidator) error {

	// Get Tendermint config base on chain id
	config := utils.GetTendermintConfig(chainId, ctx)

	// Save the KeyStore File (Optional)
	if len(keyJson) > 0 {
		keystoreDir := config.GetString("keystore")
		keyJsonFilePath := filepath.Join(keystoreDir, keystore.KeyFileName(validator.Address))
		saveKeyError := keystore.WriteKeyStore(keyJsonFilePath, keyJson)
		if saveKeyError != nil {
			return saveKeyError
		}
	}

	// Save the Validator Json File
	privValFile := config.GetString("priv_validator_file_root")
	validator.SetFile(privValFile + ".json")
	validator.Save()

	// Init the INT Genesis
	err := initEthGenesisFromExistValidator(chainId, config, validators)
	if err != nil {
		return err
	}

	// Init the INT Blockchain
	init_int_blockchain(chainId, config.GetString("int_genesis_file"), ctx)

	// Init the Tendermint Genesis
	init_em_files(config, chainId, config.GetString("int_genesis_file"), validators)

	return nil
}
