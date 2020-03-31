package ipbft

import (
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	tmcfg "github.com/intfoundation/intchain/consensus/ipbft/config/tendermint"
	cfg "github.com/intfoundation/go-config"
)

func GetTendermintConfig(chainId string, ctx *cli.Context) cfg.Config {
	datadir := ctx.GlobalString(DataDirFlag.Name)
	config := tmcfg.GetConfig(datadir, chainId)

	return config
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "intchain")
		} else {
			return filepath.Join(home, ".intchain")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func ConcatCopyPreAllocate(slices [][]byte) []byte {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]byte, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}
