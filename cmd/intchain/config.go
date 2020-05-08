// Copyright 2017 The go-ethereum Authors
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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/log"
	"io"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/intfoundation/intchain/cmd/utils"
	"github.com/intfoundation/intchain/intprotocol"
	"github.com/intfoundation/intchain/node"
	"github.com/intfoundation/intchain/params"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Eth      intprotocol.Config
	Node     node.Config
	Ethstats ethstatsConfig
}

func loadConfig(file string, cfg *gethConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "int", "eth")
	cfg.WSModules = append(cfg.WSModules, "int", "eth")
	cfg.IPCPath = "intchain.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context, chainId string) (*node.Node, gethConfig) {
	// Load defaults.
	cfg := gethConfig{
		Eth:  intprotocol.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	cfg.Node.ChainId = chainId

	// Setup Log
	//logDir := path.Join(ctx.GlobalString("datadir"), ctx.GlobalString("logDir"), chainId)
	cfg.Node.Logger = log.NewLogger(chainId, "", ctx.GlobalInt("verbosity"), ctx.GlobalBool("debug"), ctx.GlobalString("vmodule"), ctx.GlobalString("backtrace"))

	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetEthConfig(ctx, stack, &cfg.Eth)
	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ethstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}

	//utils.SetShhConfig(ctx, stack, &cfg.Shh)
	utils.SetGeneralConfig(ctx)

	return stack, cfg
}

func makeFullNode(ctx *cli.Context, cch core.CrossChainHelper, chainId string) *node.Node {
	stack, cfg := makeConfigNode(ctx, chainId)

	utils.RegisterEthService(stack, &cfg.Eth, ctx, cch)

	// Whisper must be explicitly enabled by specifying at least 1 whisper flag or in dev mode
	//shhEnabled := enableWhisper(ctx)
	//shhAutoEnabled := !ctx.GlobalIsSet(utils.WhisperEnabledFlag.Name) && ctx.GlobalIsSet(utils.DeveloperFlag.Name)
	//if shhEnabled || shhAutoEnabled {
	//	if ctx.GlobalIsSet(utils.WhisperMaxMessageSizeFlag.Name) {
	//		cfg.Shh.MaxMessageSize = uint32(ctx.Int(utils.WhisperMaxMessageSizeFlag.Name))
	//	}
	//	if ctx.GlobalIsSet(utils.WhisperMinPOWFlag.Name) {
	//		cfg.Shh.MinimumAcceptedPOW = ctx.Float64(utils.WhisperMinPOWFlag.Name)
	//	}
	//	utils.RegisterShhService(stack, &cfg.Shh)
	//}

	// Add the Ethereum Stats daemon if requested.
	//if cfg.Ethstats.URL != "" {
	//	utils.RegisterEthStatsService(stack, cfg.Ethstats.URL)
	//}
	if err := stack.GatherServices(); err != nil {
		return nil
	} else {
		return stack
	}
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx, clientIdentifier)
	comment := ""

	if cfg.Eth.Genesis != nil {
		cfg.Eth.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
