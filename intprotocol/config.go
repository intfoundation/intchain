// Copyright 2017 The go-ethereum Authors
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

package intprotocol

import (
	"math/big"
	"os"
	"os/user"
	//"path/filepath"
	"runtime"
	"time"

	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/consensus/ipbft"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/intprotocol/downloader"
	"github.com/intfoundation/intchain/intprotocol/gasprice"
	"github.com/intfoundation/intchain/params"
)

// DefaultConfig contains default settings for use on the INT Chain main net.
var DefaultConfig = Config{
	//SyncMode: downloader.FastSync,
	SyncMode:       downloader.FullSync,
	NetworkId:      1024,
	DatabaseCache:  768,
	TrieCleanCache: 256,
	TrieDirtyCache: 256,
	TrieTimeout:    60 * time.Minute,
	MinerGasFloor:  80000000,
	MinerGasCeil:   80000000,
	MinerGasPrice:  big.NewInt(params.GWei),

	TxPool: core.DefaultTxPoolConfig,
	GPO: gasprice.Config{
		Blocks:     20,
		Percentile: 60,
	},
}

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if user, err := user.Current(); err == nil {
			home = user.HomeDir
		}
	}
	if runtime.GOOS == "windows" {
		//DefaultConfig.DatasetDir = filepath.Join(home, "AppData", "Ethash")
	} else {
		//DefaultConfig.Ethash.DatasetDir = filepath.Join(home, ".ethash")
	}
}

//go:generate gencodec -type Config -field-override configMarshaling -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ethereum main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode

	NoPruning bool // Whether to disable pruning and flush everything to disk

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration

	// Mining-related options
	Coinbase      common.Address `toml:",omitempty"`
	ExtraData     []byte         `toml:",omitempty"`
	MinerGasFloor uint64
	MinerGasCeil  uint64
	MinerGasPrice *big.Int

	// Solidity compiler path
	SolcPath string

	// Transaction pool options
	TxPool core.TxPoolConfig

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Tendermint options
	IPBFT ipbft.Config

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// Data Reduction options
	PruneStateData bool
	PruneBlockData bool
}

type configMarshaling struct {
	ExtraData hexutil.Bytes
}
