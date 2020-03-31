package chain

import (
	"github.com/intfoundation/intchain/cmd/utils"
	tmcfg "github.com/intfoundation/intchain/consensus/ipbft/config/tendermint"
	cfg "github.com/intfoundation/go-config"
	"gopkg.in/urfave/cli.v1"
)

var (
	// ipbft config
	Config cfg.Config
)

func GetTendermintConfig(chainId string, ctx *cli.Context) cfg.Config {
	datadir := ctx.GlobalString(utils.DataDirFlag.Name)
	config := tmcfg.GetConfig(datadir, chainId)

	return config
}

func contains(a []string, s string) bool {
	for _, e := range a {
		if s == e {
			return true
		}
	}
	return false
}
