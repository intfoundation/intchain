package ethapi

import (
	"context"
	"errors"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft/epoch"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/state"
	"github.com/intfoundation/intchain/core/types"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/rpc"
	"math/big"
)

type PublicDelegateAPI struct {
	b Backend
}

func NewPublicDelegateAPI(b Backend) *PublicDelegateAPI {
	return &PublicDelegateAPI{
		b: b,
	}
}

var (
	defaultSelfSecurityDeposit = math.MustParseBig256("10000000000000000000000") // 10,000 * e18
	minimumDelegationAmount    = math.MustParseBig256("1000000000000000000000")  // 1000 * e18

	maxDelegationAddresses = 1000
)

func (api *PublicDelegateAPI) Delegate(ctx context.Context, from, candidate common.Address, amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.Delegate.String(), candidate)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.Delegate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    amount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (api *PublicDelegateAPI) CancelDelegate(ctx context.Context, from, candidate common.Address, amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.CancelDelegate.String(), candidate, (*big.Int)(amount))
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.CancelDelegate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (api *PublicDelegateAPI) ApplyCandidate(ctx context.Context, from common.Address, securityDeposit *hexutil.Big, commission uint8, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.Candidate.String(), commission)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.Candidate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    securityDeposit,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (api *PublicDelegateAPI) CancelCandidate(ctx context.Context, from common.Address, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.CancelCandidate.String())
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.CancelCandidate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func (api *PublicDelegateAPI) CheckCandidate(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (map[string]interface{}, error) {
	state, _, err := api.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}

	//fmt.Printf("ethapi delegate api CheckCandidate address %v\n", address)

	fields := map[string]interface{}{
		"candidate":  state.IsCandidate(address),
		"commission": state.GetCommission(address),
	}
	return fields, state.Error()
}

func (api *PublicDelegateAPI) SetCommission(ctx context.Context, from common.Address, commission uint8, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.SetCommission.String(), commission)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.SetCommission.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return api.b.GetInnerAPIBridge().SendTransaction(ctx, args)
}

func init() {
	// Delegate
	core.RegisterValidateCb(intAbi.Delegate, del_ValidateCb)
	core.RegisterApplyCb(intAbi.Delegate, del_ApplyCb)

	// Cancel Delegate
	core.RegisterValidateCb(intAbi.CancelDelegate, cdel_ValidateCb)
	core.RegisterApplyCb(intAbi.CancelDelegate, cdel_ApplyCb)

	// Candidate
	core.RegisterValidateCb(intAbi.Candidate, appcdd_ValidateCb)
	core.RegisterApplyCb(intAbi.Candidate, appcdd_ApplyCb)

	// Cancel Candidate
	core.RegisterValidateCb(intAbi.CancelCandidate, ccdd_ValidateCb)
	core.RegisterApplyCb(intAbi.CancelCandidate, ccdd_ApplyCb)

	// Set Commission
	core.RegisterValidateCb(intAbi.SetCommission, setcom_ValidateCb)
	core.RegisterApplyCb(intAbi.SetCommission, setcom_ApplyCb)
}

func del_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := delegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func del_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := delegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// Do job
	amount := tx.Value()
	// Move Balance to delegate balance
	state.SubBalance(from, amount)
	state.AddDelegateBalance(from, amount)
	// Add Balance to Candidate's Proxied Balance
	state.AddProxiedBalanceByUser(args.Candidate, from, amount)

	return nil
}

func cdel_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := cancelDelegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func cdel_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := cancelDelegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// Apply Logic
	// if request amount < proxied amount, refund it immediately
	// otherwise, refund the proxied amount, and put the rest to pending refund balance
	proxiedBalance := state.GetProxiedBalanceByUser(args.Candidate, from)
	var immediatelyRefund *big.Int
	if args.Amount.Cmp(proxiedBalance) <= 0 {
		immediatelyRefund = args.Amount
	} else {
		immediatelyRefund = proxiedBalance
		restRefund := new(big.Int).Sub(args.Amount, proxiedBalance)
		state.AddPendingRefundBalanceByUser(args.Candidate, from, restRefund)
		// TODO Add Pending Refund Set, Commit the Refund Set
		state.MarkDelegateAddressRefund(args.Candidate)
	}

	state.SubProxiedBalanceByUser(args.Candidate, from, immediatelyRefund)
	state.SubDelegateBalance(from, immediatelyRefund)
	state.AddBalance(from, immediatelyRefund)

	return nil
}

func appcdd_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := candidateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func appcdd_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := candidateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	amount := tx.Value()
	// Add security deposit to self
	state.SubBalance(from, amount)
	state.AddDelegateBalance(from, amount)
	state.AddProxiedBalanceByUser(from, from, amount)
	// Become a Candidate
	state.ApplyForCandidate(from, args.Commission)

	// mark address candidate
	state.MarkAddressCandidate(from)

	return nil
}

func ccdd_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	verror := cancelCandidateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func ccdd_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	verror := cancelCandidateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// Do job
	allRefund := true
	// Refund all the amount back to users
	state.ForEachProxied(from, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
		// Refund Proxied Amount
		state.SubProxiedBalanceByUser(from, key, proxiedBalance)
		state.SubDelegateBalance(key, proxiedBalance)
		state.AddBalance(key, proxiedBalance)

		if depositProxiedBalance.Sign() > 0 {
			allRefund = false
			// Refund Deposit to PendingRefund if deposit > 0
			state.AddPendingRefundBalanceByUser(from, key, depositProxiedBalance)
			// TODO Add Pending Refund Set, Commit the Refund Set
			state.MarkDelegateAddressRefund(from)
		}
		return true
	})

	state.CancelCandidate(from, allRefund)

	// remove address form candidate set
	state.ClearCandidateSetByAddress(from)

	return nil
}

// set commission
func setcom_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, err := setCommissionValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func setcom_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	from := derivedAddressFromTx(tx)
	args, err := setCommissionValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	state.SetCommission(from, args.Commission)

	return nil
}

// Validation

func delegateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.DelegateArgs, error) {
	// Check minimum delegate amount
	if tx.Value().Cmp(minimumDelegationAmount) < 0 {
		return nil, core.ErrDelegateAmount
	}

	var args intAbi.DelegateArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.Delegate.String(), data[4:]); err != nil {
		return nil, err
	}

	// Check Candidate
	if !state.IsCandidate(args.Candidate) {
		return nil, core.ErrNotCandidate
	}

	depositBalance := state.GetDepositProxiedBalanceByUser(args.Candidate, from)
	if depositBalance.Sign() == 0 {
		// Check if exceed the limit of delegated addresses
		// if exceed the limit of delegation address number, return error
		delegatedAddressNumber := state.GetProxiedAddressNumber(args.Candidate)
		if delegatedAddressNumber >= maxDelegationAddresses {
			return nil, core.ErrExceedDelegationAddressLimit
		}
	}

	// If Candidate is supernode, only allow to increase the stack(whitelist proxied list), not allow to create the new stack
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(args.Candidate.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		if depositBalance.Sign() == 0 {
			return nil, core.ErrCannotDelegate
		}
	}

	// Check Epoch Height
	if err := checkEpochInNormalStage(bc); err != nil {
		return nil, err
	}
	return &args, nil
}

func cancelDelegateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.CancelDelegateArgs, error) {

	var args intAbi.CancelDelegateArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.CancelDelegate.String(), data[4:]); err != nil {
		return nil, err
	}

	// Check Self Address
	if from == args.Candidate {
		return nil, core.ErrCancelSelfDelegate
	}

	// Super node Candidate can't decrease balance
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(args.Candidate.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return nil, core.ErrCannotCancelDelegate
	}

	// Check Proxied Amount in Candidate Balance
	proxiedBalance := state.GetProxiedBalanceByUser(args.Candidate, from)
	depositProxiedBalance := state.GetDepositProxiedBalanceByUser(args.Candidate, from)
	pendingRefundBalance := state.GetPendingRefundBalanceByUser(args.Candidate, from)
	// net = deposit - pending refund
	netDeposit := new(big.Int).Sub(depositProxiedBalance, pendingRefundBalance)
	// available = proxied + net
	availableRefundBalance := new(big.Int).Add(proxiedBalance, netDeposit)
	if args.Amount.Cmp(availableRefundBalance) == 1 {
		return nil, core.ErrInsufficientProxiedBalance
	}

	remainingBalance := new(big.Int).Sub(availableRefundBalance, args.Amount)
	if remainingBalance.Sign() == 1 && remainingBalance.Cmp(minimumDelegationAmount) == -1 {
		return nil, core.ErrDelegateAmount
	}

	// Check Epoch Height
	if err := checkEpochInNormalStage(bc); err != nil {
		return nil, err
	}

	return &args, nil
}

func candidateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.CandidateArgs, error) {
	// Check cleaned Candidate
	if !state.IsCleanAddress(from) {
		return nil, core.ErrAlreadyCandidate
	}

	// Check minimum Security Deposit
	if tx.Value().Cmp(defaultSelfSecurityDeposit) == -1 {
		return nil, core.ErrMinimumSecurityDeposit
	}

	var args intAbi.CandidateArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.Candidate.String(), data[4:]); err != nil {
		return nil, err
	}

	// Check Commission Range
	if args.Commission > 100 {
		return nil, core.ErrCommission
	}

	// Check Epoch Height
	if err := checkEpochInNormalStage(bc); err != nil {
		return nil, err
	}

	// Annual/SemiAnnual supernode can not become candidate
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(from.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return nil, core.ErrCannotCandidate
	}

	return &args, nil
}

func cancelCandidateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	// Check already Candidate
	if !state.IsCandidate(from) {
		return core.ErrNotCandidate
	}

	// Super node can't cancel Candidate
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(from.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return core.ErrCannotCancelCandidate
	}

	// Check Epoch Height
	if err := checkEpochInNormalStage(bc); err != nil {
		return err
	}

	return nil
}

func setCommissionValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.SetCommissionArgs, error) {
	if !state.IsCandidate(from) {
		return nil, core.ErrNotCandidate
	}

	var args intAbi.SetCommissionArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.SetCommission.String(), data[4:]); err != nil {
		return nil, err
	}

	if args.Commission > 100 {
		return nil, core.ErrCommission
	}

	return &args, nil
}

// Common
func derivedAddressFromTx(tx *types.Transaction) (from common.Address) {
	signer := types.NewEIP155Signer(tx.ChainId())
	from, _ = types.Sender(signer, tx)
	return
}

func checkEpochInNormalStage(bc *core.BlockChain) error {
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}

	if ep == nil {
		return errors.New("epoch is nil, are you running on IPBFT Consensus Engine")
	}

	// Vote is valid between height 0% - 75%
	//height := bc.CurrentBlock().NumberU64()
	//if !ep.CheckInNormalStage(height) {
	//	return errors.New(fmt.Sprintf("you can't send this tx during this time, current height %v", height))
	//}
	return nil
}
