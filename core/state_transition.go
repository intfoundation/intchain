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

package core

import (
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/log"
	"math"
	"math/big"

	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/core/vm"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/params"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

// ExecutionResult includes all output after executing given evm
// message no matter the execution itself is successful or not.
type ExecutionResult struct {
	UsedGas    uint64 // Total used gas but include the refunded gas
	Err        error  // Any error encountered during the execution(listed in core/vm/errors.go)
	ReturnData []byte // Returned data from evm(function result or data supplied with revert opcode)
}

// Unwrap returns the internal evm error which allows us for further
// analysis outside.
func (result *ExecutionResult) Unwrap() error {
	return result.Err
}

// Failed returns the indicator whether the execution is successful or not
func (result *ExecutionResult) Failed() bool { return result.Err != nil }

// Return is a helper function to help caller distinguish between revert reason
// and function return. Return returns the data after execution if no error occurs.
func (result *ExecutionResult) Return() []byte {
	if result.Err != nil {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// Revert returns the concrete revert reason if the execution is aborted by `REVERT`
// opcode. Note the reason can be nil if no data supplied with revert opcode.
func (result *ExecutionResult) Revert() []byte {
	if result.Err != vm.ErrExecutionReverted {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation, homestead bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation && homestead {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		nonZeroGas := params.TxDataNonZeroGasFrontier
		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * nonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		value:    msg.Value(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) (*ExecutionResult, error) {
	return NewStateTransition(evm, msg, gp).TransitionDb()
}

func (st *StateTransition) from() vm.AccountRef {
	f := st.msg.From()
	if !st.state.Exist(f) {
		st.state.CreateAccount(f)
	}
	return vm.AccountRef(f)
}

func (st *StateTransition) to() vm.AccountRef {
	if st.msg == nil {
		return vm.AccountRef{}
	}
	to := st.msg.To()
	if to == nil {
		return vm.AccountRef{} // contract creation
	}

	reference := vm.AccountRef(*to)
	if !st.state.Exist(*to) {
		st.state.CreateAccount(*to)
	}
	return reference
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	var (
		state  = st.state
		sender = st.from()
	)
	mgval := new(big.Int).Mul(new(big.Int).SetUint64(st.msg.Gas()), st.gasPrice)
	if state.GetBalance(sender.Address()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	state.SubBalance(sender.Address(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	msg := st.msg
	sender := st.from()

	// Make sure this transaction's nonce is correct
	if msg.CheckNonce() {
		nonce := st.state.GetNonce(sender.Address())
		if nonce < msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.buyGas()
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the the used gas. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (*ExecutionResult, error) {
	if err := st.preCheck(); err != nil {
		return nil, err
	}
	msg := st.msg
	sender := st.from() // err checked in preCheck

	homestead := st.evm.ChainConfig().IsHomestead(st.evm.BlockNumber)
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation, homestead)
	if err != nil {
		return nil, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, err
	}

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
		ret   []byte
	)
	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(sender.Address(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.to().Address(), st.data, st.gas, st.value)
	}
	//if vmerr != nil {
	//	log.Debug("VM returned with error", "err", vmerr)
	//	// The only possible consensus-error would be if there wasn't
	//	// sufficient balance to make the transfer happen. The first
	//	// balance transfer may never fail.
	//	if vmerr == vm.ErrInsufficientBalance {
	//		return nil, 0, false, vmerr
	//	}
	//}
	st.refundGas()
	st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	return &ExecutionResult{
		UsedGas:    st.gasUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}, nil
}

func (st *StateTransition) refundGas() {

	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	sender := st.from()

	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)

	st.state.AddBalance(sender.Address(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessageEx(evm *vm.EVM, msg Message, gp *GasPool) (*ExecutionResult, *big.Int, error) {
	return NewStateTransition(evm, msg, gp).TransitionDbEx()
}

// TransitionDbEx will move the state by applying the message against the given environment.
func (st *StateTransition) TransitionDbEx() (*ExecutionResult, *big.Int, error) {

	if err := st.preCheck(); err != nil {
		return nil, nil, err
	}
	msg := st.msg
	sender := st.from() // err checked in preCheck

	homestead := st.evm.ChainConfig().IsHomestead(st.evm.BlockNumber)
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation, homestead)
	if err != nil {
		return nil, nil, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, nil, err
	}

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
		ret   []byte
	)

	//log.Debugf("TransitionDbEx 0\n")

	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		//log.Debugf("TransitionDbEx 1, sender is %x, nonce is \n", sender.Address(), st.state.GetNonce(sender.Address())+1)

		st.state.SetNonce(sender.Address(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.to().Address(), st.data, st.gas, st.value)

		//log.Debugf("TransitionDbEx 2\n")

	}

	//log.Debugf("TransitionDbEx 3\n")

	//if vmerr != nil {
	//	log.Debug("VM returned with error", "err", vmerr)
	//	// The only possible consensus-error would be if there wasn't
	//	// sufficient balance to make the transfer happen. The first
	//	// balance transfer may never fail.
	//	if vmerr == vm.ErrInsufficientBalance {
	//		return nil, 0, nil, false, vmerr
	//	}
	//}

	st.refundGas()

	//log.Debugf("TransitionDbEx 4, coinbase is %x, balance is %v\n",
	//	st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	//st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))
	usedMoney := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)

	//log.Debugf("TransitionDbEx 5\n")

	//log.Debugf("TransitionDbEx, send.balance is %v\n", st.state.GetBalance(sender.Address()))
	//log.Debugf("TransitionDbEx, return ret-%v, st.gasUsed()-%v, usedMoney-%v, vmerr-%v, err-%v\n",
	//	ret, st.gasUsed(), usedMoney, vmerr, err)

	//return ret, st.gasUsed(), usedMoney, vmerr != nil, err

	return &ExecutionResult{
		UsedGas:    st.gasUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}, usedMoney, nil
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessageTracer(evm *vm.EVM, msg Message, gp *GasPool) (*ExecutionResult, *big.Int, error) {
	return NewStateTransition(evm, msg, gp).TransitionDbTracer()
}

// TransitionDbEx will move the state by applying the message against the given environment.
func (st *StateTransition) TransitionDbTracer() (*ExecutionResult, *big.Int, error) {

	if err := st.preCheck(); err != nil {
		return nil, nil, err
	}
	msg := st.msg
	sender := st.from() // err checked in preCheck

	homestead := st.evm.ChainConfig().IsHomestead(st.evm.BlockNumber)
	contractCreation := msg.To() == nil

	if !contractCreation && intAbi.IsIntChainContractAddr(msg.To()) {
		data := st.data
		from := msg.From()
		function, err := intAbi.FunctionTypeFromId(data[:4])
		if err != nil {
			return nil, nil, err
		}

		if msg.CheckNonce() {
			nonce := st.state.GetNonce(from)
			if nonce < msg.Nonce() {
				log.Info("ApplyTransactionEx() abort due to nonce too high")
				return nil, nil, ErrNonceTooHigh
			} else if nonce > msg.Nonce() {
				log.Info("ApplyTransactionEx() abort due to nonce too low")
				return nil, nil, ErrNonceTooLow
			}
		}

		// pre-buy gas according to the gas limit
		gasLimit := msg.Gas()
		gasValue := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), msg.GasPrice())
		if st.state.GetBalance(from).Cmp(gasValue) < 0 {
			fmt.Printf("insufficient INT for gas (%x). Req %v, has %v", from.Bytes()[:4], gasValue, st.state.GetBalance(from))
			//return nil, nil, fmt.Errorf("insufficient INT for gas (%x). Req %v, has %v", from.Bytes()[:4], gasValue, st.state.GetBalance(from))
		}

		//if err := st.gp.SubGas(gasLimit); err != nil {
		//	return nil, nil, err
		//}

		//st.state.SubBalance(from, gasValue)
		//log.Infof("ApplyTransactionEx() 1, gas is %v, gasPrice is %v, gasValue is %v\n", gasLimit, tx.GasPrice(), gasValue)

		// use gas
		gas := function.RequiredGas()
		if gasLimit < gas {
			fmt.Printf("out of gas, req %v, has %v\n", gas, gasLimit)
			return nil, nil, vm.ErrOutOfGas
		}

		restBalance := new(big.Int).Sub(st.state.GetBalance(from), gasValue)

		// Check Tx Amount
		if restBalance.Cmp(msg.Value()) == -1 {
			fmt.Printf("insufficient INT for tx amount (%x). Req %v, has %v", from.Bytes()[:4], msg.Value(), restBalance)
			//return nil, nil, fmt.Errorf("insufficient INT for tx amount (%x). Req %v, has %v", from.Bytes()[:4], msg.Value(), restBalance)
		}

		//if applyCb := GetApplyCb(function); applyCb != nil {
		//	if fn, ok := applyCb.(NonCrossChainApplyCb); ok {
		//		if err := fn(msg, st.state., bc, ops); err != nil {
		//			return nil, nil, err
		//		}
		//	} else {
		//		panic("callback func is wrong, this should not happened, please check the code")
		//	}
		//}

		usedMoney := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)

		return &ExecutionResult{
			UsedGas:    st.gasUsed(),
			Err:        nil,
			ReturnData: []byte{1},
		}, usedMoney, nil

	}

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation, homestead)
	if err != nil {
		return nil, nil, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, nil, err
	}

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
		ret   []byte
	)

	//log.Debugf("TransitionDbEx 0\n")

	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		//log.Debugf("TransitionDbEx 1, sender is %x, nonce is \n", sender.Address(), st.state.GetNonce(sender.Address())+1)

		st.state.SetNonce(sender.Address(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.to().Address(), st.data, st.gas, st.value)

		//log.Debugf("TransitionDbEx 2\n")

	}

	//log.Debugf("TransitionDbEx 3\n")

	//if vmerr != nil {
	//	log.Debug("VM returned with error", "err", vmerr)
	//	// The only possible consensus-error would be if there wasn't
	//	// sufficient balance to make the transfer happen. The first
	//	// balance transfer may never fail.
	//	if vmerr == vm.ErrInsufficientBalance {
	//		return nil, 0, nil, false, vmerr
	//	}
	//}

	st.refundGas()

	//log.Debugf("TransitionDbEx 4, coinbase is %x, balance is %v\n",
	//	st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))

	//st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice))
	usedMoney := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)

	//log.Debugf("TransitionDbEx 5\n")

	//log.Debugf("TransitionDbEx, send.balance is %v\n", st.state.GetBalance(sender.Address()))
	//log.Debugf("TransitionDbEx, return ret-%v, st.gasUsed()-%v, usedMoney-%v, vmerr-%v, err-%v\n",
	//	ret, st.gasUsed(), usedMoney, vmerr, err)

	//return ret, st.gasUsed(), usedMoney, vmerr != nil, err

	fmt.Printf("ret %X\n", ret)
	return &ExecutionResult{
		UsedGas:    st.gasUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}, usedMoney, nil
}
