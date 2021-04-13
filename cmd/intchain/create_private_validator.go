package main

import (
	"fmt"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/go-wire"
	"github.com/intfoundation/intchain/cmd/utils"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/log"
	"github.com/intfoundation/intchain/params"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
)

type PrivValidatorForConsole struct {
	// IntChain Account Address
	Address string `json:"address"`
	// IntChain Consensus Public Key, in BLS format
	PubKey crypto.PubKey `json:"consensus_pub_key"`
	// IntChain Consensus Private Key, in BLS format
	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey crypto.PrivKey `json:"consensus_priv_key"`
}

func CreatePrivateValidatorCmd(ctx *cli.Context) error {
	address := ctx.Args().First()

	if address == "" {
		log.Info("address is empty, need an address")
		return nil
	}

	datadir := ctx.GlobalString(utils.DataDirFlag.Name)
	if err := os.MkdirAll(datadir, 0700); err != nil {
		return err
	}

	chainId := params.MainnetChainConfig.IntChainId

	if ctx.GlobalIsSet(utils.TestnetFlag.Name) {
		chainId = params.TestnetChainConfig.IntChainId
	}

	privValFilePath := filepath.Join(ctx.GlobalString(utils.DataDirFlag.Name), chainId)
	privValFile := filepath.Join(ctx.GlobalString(utils.DataDirFlag.Name), chainId, "priv_validator.json")

	err := os.MkdirAll(privValFilePath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	validator := types.GenPrivValidatorKey(common.HexToAddress(address))

	fmt.Printf(string(wire.JSONBytesPretty(validator)))
	validator.SetFile(privValFile)
	validator.Save()

	return nil
}
