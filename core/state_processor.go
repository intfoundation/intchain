// Copyright 2015 The go-ethereum Authors
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

package core

import (
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/core/state"
	"github.com/intfoundation/intchain/core/types"
	"github.com/intfoundation/intchain/core/vm"
	"github.com/intfoundation/intchain/crypto"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/log"
	"github.com/intfoundation/intchain/params"
	"math/big"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
	cch    CrossChainHelper
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine, cch CrossChainHelper) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
		cch:    cch,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, *types.PendingOps, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
		ops      = new(types.PendingOps)
	)
	// Mutate the the block and state according to any hard-fork specs
	//if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
	//	misc.ApplyDAOHardFork(statedb)
	//}
	totalUsedMoney := big.NewInt(0)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		//receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		receipt, _, err := ApplyTransactionEx(p.config, p.bc, nil, gp, statedb, ops, header, tx,
			usedGas, totalUsedMoney, cfg, p.cch, false)
		log.Debugf("(p *StateProcessor) Process()，after ApplyTransactionEx, receipt is %v\n", receipt)
		if err != nil {
			return nil, nil, 0, nil, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	_, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), totalUsedMoney, block.Uncles(), receipts, ops)
	if err != nil {
		return nil, nil, 0, nil, err
	}

	return receipts, allLogs, *usedGas, ops, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc *BlockChain, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, err
	}
	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, 0, err
	}
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}

// ApplyTransactionEx attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransactionEx(config *params.ChainConfig, bc *BlockChain, author *common.Address, gp *GasPool, statedb *state.StateDB, ops *types.PendingOps,
	header *types.Header, tx *types.Transaction, usedGas *uint64, totalUsedMoney *big.Int, cfg vm.Config, cch CrossChainHelper, mining bool) (*types.Receipt, uint64, error) {

	signer := types.MakeSigner(config, header.Number)
	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, 0, err
	}

	if !intAbi.IsIntChainContractAddr(tx.To()) {

		//log.Debugf("ApplyTransactionEx 1\n")

		// Create a new context to be used in the EVM environment
		context := NewEVMContext(msg, header, bc, author)

		//log.Debugf("ApplyTransactionEx 2\n")

		// Create a new environment which holds all relevant information
		// about the transaction and calling mechanisms.
		vmenv := vm.NewEVM(context, statedb, config, cfg)
		// Apply the transaction to the current state (included in the env)
		_, gas, money, failed, err := ApplyMessageEx(vmenv, msg, gp)
		if err != nil {
			return nil, 0, err
		}

		//log.Debugf("ApplyTransactionEx 3\n")
		// Update the state with pending changes
		var root []byte
		if config.IsByzantium(header.Number) {
			//log.Debugf("ApplyTransactionEx(), is byzantium\n")
			statedb.Finalise(true)
		} else {
			//log.Debugf("ApplyTransactionEx(), is not byzantium\n")
			root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
		}
		*usedGas += gas
		totalUsedMoney.Add(totalUsedMoney, money)

		// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
		// based on the eip phase, we're passing wether the root touch-delete accounts.
		receipt := types.NewReceipt(root, failed, *usedGas)
		log.Debugf("ApplyTransactionEx，new receipt with (root,failed,*usedGas) = (%v,%v,%v)\n", root, failed, *usedGas)
		receipt.TxHash = tx.Hash()
		//log.Debugf("ApplyTransactionEx，new receipt with txhash %v\n", receipt.TxHash)
		receipt.GasUsed = gas
		//log.Debugf("ApplyTransactionEx，new receipt with gas %v\n", receipt.GasUsed)
		// if the transaction created a contract, store the creation address in the receipt.
		if msg.To() == nil {
			receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
		}
		// Set the receipt logs and create a bloom for filtering
		receipt.Logs = statedb.GetLogs(tx.Hash())
		//log.Debugf("ApplyTransactionEx，new receipt with receipt.Logs %v\n", receipt.Logs)
		receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
		receipt.BlockHash = statedb.BlockHash()
		receipt.BlockNumber = header.Number
		receipt.TransactionIndex = uint(statedb.TxIndex())
		//log.Debugf("ApplyTransactionEx，new receipt with receipt.Bloom %v\n", receipt.Bloom)
		//log.Debugf("ApplyTransactionEx 4\n")
		return receipt, gas, err

	} else {

		// the first 4 bytes is the function identifier
		data := tx.Data()
		function, err := intAbi.FunctionTypeFromId(data[:4])
		if err != nil {
			return nil, 0, err
		}
		log.Infof("ApplyTransactionEx() 0, Chain Function is %v", function.String())

		// check Function main/child flag
		if config.IsMainChain() && !function.AllowInMainChain() {
			return nil, 0, ErrNotAllowedInMainChain
		} else if !config.IsMainChain() && !function.AllowInChildChain() {
			return nil, 0, ErrNotAllowedInChildChain
		}

		from := msg.From()
		// Make sure this transaction's nonce is correct
		if msg.CheckNonce() {
			nonce := statedb.GetNonce(from)
			if nonce < msg.Nonce() {
				log.Info("ApplyTransactionEx() abort due to nonce too high")
				return nil, 0, ErrNonceTooHigh
			} else if nonce > msg.Nonce() {
				log.Info("ApplyTransactionEx() abort due to nonce too low")
				return nil, 0, ErrNonceTooLow
			}
		}

		// pre-buy gas according to the gas limit
		gasLimit := tx.Gas()
		gasValue := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), tx.GasPrice())
		if statedb.GetBalance(from).Cmp(gasValue) < 0 {
			return nil, 0, fmt.Errorf("insufficient INT for gas (%x). Req %v, has %v", from.Bytes()[:4], gasValue, statedb.GetBalance(from))
		}
		if err := gp.SubGas(gasLimit); err != nil {
			return nil, 0, err
		}
		statedb.SubBalance(from, gasValue)
		//log.Infof("ApplyTransactionEx() 1, gas is %v, gasPrice is %v, gasValue is %v\n", gasLimit, tx.GasPrice(), gasValue)

		// use gas
		gas := function.RequiredGas()
		if gasLimit < gas {
			return nil, 0, vm.ErrOutOfGas
		}

		// Check Tx Amount
		if statedb.GetBalance(from).Cmp(tx.Value()) == -1 {
			return nil, 0, fmt.Errorf("insufficient INT for tx amount (%x). Req %v, has %v", from.Bytes()[:4], tx.Value(), statedb.GetBalance(from))
		}

		if applyCb := GetApplyCb(function); applyCb != nil {
			if function.IsCrossChainType() {
				if fn, ok := applyCb.(CrossChainApplyCb); ok {
					cch.GetMutex().Lock()
					err := fn(tx, statedb, ops, cch, mining)
					cch.GetMutex().Unlock()

					if err != nil {
						return nil, 0, err
					}
				} else {
					panic("callback func is wrong, this should not happened, please check the code")
				}
			} else {
				if fn, ok := applyCb.(NonCrossChainApplyCb); ok {
					if err := fn(tx, statedb, bc, ops); err != nil {
						return nil, 0, err
					}
				} else {
					panic("callback func is wrong, this should not happened, please check the code")
				}
			}
		}

		// refund gas
		remainingGas := gasLimit - gas
		remaining := new(big.Int).Mul(new(big.Int).SetUint64(remainingGas), tx.GasPrice())
		statedb.AddBalance(from, remaining)
		gp.AddGas(remainingGas)

		*usedGas += gas
		totalUsedMoney.Add(totalUsedMoney, new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice()))
		log.Infof("ApplyTransactionEx() 2, totalUsedMoney is %v\n", totalUsedMoney)

		// Update the state with pending changes
		var root []byte
		if config.IsByzantium(header.Number) {
			statedb.Finalise(true)
		} else {
			root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
		}

		// TODO: whether true of false for the transaction receipt
		receipt := types.NewReceipt(root, false, *usedGas)
		receipt.TxHash = tx.Hash()
		receipt.GasUsed = gas

		// Set the receipt logs and create a bloom for filtering
		receipt.Logs = statedb.GetLogs(tx.Hash())
		receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
		receipt.BlockHash = statedb.BlockHash()
		receipt.BlockNumber = header.Number
		receipt.TransactionIndex = uint(statedb.TxIndex())

		statedb.SetNonce(msg.From(), statedb.GetNonce(msg.From())+1)
		//log.Infof("ApplyTransactionEx() 3, totalUsedMoney is %v\n", totalUsedMoney)

		return receipt, 0, nil
	}
}
