package ethapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft/epoch"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/state"
	"github.com/intfoundation/intchain/core/types"
	intcrypto "github.com/intfoundation/intchain/crypto"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/go-crypto"
	"math/big"
	"time"
)

type PublicTdmAPI struct {
	b Backend
}

func NewPublicTdmAPI(b Backend) *PublicTdmAPI {
	return &PublicTdmAPI{
		b: b,
	}
}

var (
	minimumVoteAmount      = math.MustParseBig256("100000000000000000000000") // 100,000 * e18
	maxEditValidatorLength = 100
)

func (api *PublicTdmAPI) VoteNextEpoch(ctx context.Context, from common.Address, voteHash common.Hash, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.VoteNextEpoch.String(), voteHash)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.VoteNextEpoch.RequiredGas()

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

func (api *PublicTdmAPI) RevealVote(ctx context.Context, from common.Address, pubkey crypto.BLSPubKey, amount *hexutil.Big, salt string, signature hexutil.Bytes, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.RevealVote.String(), pubkey.Bytes(), (*big.Int)(amount), salt, signature)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.RevealVote.RequiredGas()

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

func (api *PublicTdmAPI) EditValidator(ctx context.Context, from common.Address, moniker, website string, identity string, details string, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.EditValidator.String(), moniker, website, identity, details)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.EditValidator.RequiredGas()

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

//func (api *PublicTdmAPI) GetValidatorStatus(ctx context.Context, from common.Address, blockNr rpc.BlockNumber) (map[string]interface{}, error) {
//	state, _, err := api.b.StateAndHeaderByNumber(ctx, blockNr)
//	if state == nil || err != nil {
//		return nil, err
//	}
//	fields := map[string]interface{}{
//		"IsForbidden": state.GetOrNewStateObject(from).IsForbidden(),
//	}
//
//	return fields, nil
//}

func (api *PublicTdmAPI) UnForbid(ctx context.Context, from common.Address, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.UnForbid.String())
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.UnForbid.RequiredGas()

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
	// Vote for Next Epoch
	core.RegisterValidateCb(intAbi.VoteNextEpoch, vne_ValidateCb)
	core.RegisterApplyCb(intAbi.VoteNextEpoch, vne_ApplyCb)

	// Reveal Vote
	core.RegisterValidateCb(intAbi.RevealVote, rev_ValidateCb)
	core.RegisterApplyCb(intAbi.RevealVote, rev_ApplyCb)

	// Edit Validator
	core.RegisterValidateCb(intAbi.EditValidator, edv_ValidateCb)

	// UnForbid
	core.RegisterValidateCb(intAbi.UnForbid, unf_ValidateCb)
	core.RegisterApplyCb(intAbi.UnForbid, unf_ApplyCb)

}

func vne_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {

	_, verror := voteNextEpochValidation(tx, bc)
	if verror != nil {
		return verror
	}

	return nil
}

func vne_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := voteNextEpochValidation(tx, bc)
	if verror != nil {
		return verror
	}

	op := types.VoteNextEpochOp{
		From:     from,
		VoteHash: args.VoteHash,
		TxHash:   tx.Hash(),
	}

	if ok := ops.Append(&op); !ok {
		return fmt.Errorf("pending ops conflict: %v", op)
	}

	return nil
}

func rev_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := revealVoteValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func rev_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {

	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := revealVoteValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// Apply Logic
	if state.IsCandidate(from) {
		// Move delegate amount first if Candidate
		state.ForEachProxied(from, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
			// Move Proxied Amount to Deposit Proxied Amount
			state.SubProxiedBalanceByUser(from, key, proxiedBalance)
			state.AddDepositProxiedBalanceByUser(from, key, proxiedBalance)
			return true
		})
	}

	// Rest Vote Amount
	proxiedBalance := state.GetTotalProxiedBalance(from)
	depositProxiedBalance := state.GetTotalDepositProxiedBalance(from)
	pendingRefundBalance := state.GetTotalPendingRefundBalance(from)
	netProxied := new(big.Int).Sub(new(big.Int).Add(proxiedBalance, depositProxiedBalance), pendingRefundBalance)
	netSelfAmount := new(big.Int).Sub(args.Amount, netProxied)

	// if lock balance less than net self amount, then add enough amount to locked balance
	if state.GetDepositBalance(from).Cmp(netSelfAmount) == -1 {
		difference := new(big.Int).Sub(netSelfAmount, state.GetDepositBalance(from))
		state.SubBalance(from, difference)
		state.AddDepositBalance(from, difference)
	}

	var pub crypto.BLSPubKey
	copy(pub[:], args.PubKey)

	op := types.RevealVoteOp{
		From:   from,
		Pubkey: pub,
		Amount: args.Amount,
		Salt:   args.Salt,
		TxHash: tx.Hash(),
	}

	if ok := ops.Append(&op); !ok {
		return fmt.Errorf("pending ops conflict: %v", op)
	}

	return nil
}

func edv_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	if !state.IsCandidate(from) {
		return errors.New("you are not a validator or candidate")
	}

	var args intAbi.EditValidatorArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.EditValidator.String(), data[4:]); err != nil {
		return err
	}

	if len([]byte(args.Details)) > maxEditValidatorLength ||
		len([]byte(args.Identity)) > maxEditValidatorLength ||
		len([]byte(args.Moniker)) > maxEditValidatorLength ||
		len([]byte(args.Website)) > maxEditValidatorLength {
		//fmt.Printf("args details length %v, identity length %v, moniker lenth %v, website length %v\n", len([]byte(args.Details)),len([]byte(args.Identity)),len([]byte(args.Moniker)),len([]byte(args.Website)))
		return fmt.Errorf("args length too long, more than %v", maxEditValidatorLength)
	}

	return nil
}

func unf_ValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)

	err := unForbidValidation(from, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func unf_ApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	err := unForbidValidation(from, state, bc)
	if err != nil {
		return err
	}

	state.GetOrNewStateObject(from).SetForbidden(false)

	// remove address from forbidden set
	state.ClearForbiddenSetByAddress(from)

	return nil
}

// Validation

func voteNextEpochValidation(tx *types.Transaction, bc *core.BlockChain) (*intAbi.VoteNextEpochArgs, error) {
	var args intAbi.VoteNextEpochArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.VoteNextEpoch.String(), data[4:]); err != nil {
		return nil, err
	}

	// Check Epoch Height
	if err := checkEpochInHashVoteStage(bc); err != nil {
		return nil, err
	}

	return &args, nil
}

func revealVoteValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.RevealVoteArgs, error) {
	var args intAbi.RevealVoteArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.RevealVote.String(), data[4:]); err != nil {
		return nil, err
	}
	//fmt.Printf("tdm api args %v\n", args)
	var netProxied *big.Int
	if state.IsCandidate(from) {
		// is Candidate? Check Proxied Balance (Amount >= (proxiedBalance + depositProxiedBalance - pendingRefundBalance))
		proxiedBalance := state.GetTotalProxiedBalance(from)
		depositProxiedBalance := state.GetTotalDepositProxiedBalance(from)
		pendingRefundBalance := state.GetTotalPendingRefundBalance(from)
		netProxied = new(big.Int).Sub(new(big.Int).Add(proxiedBalance, depositProxiedBalance), pendingRefundBalance)
	} else {
		netProxied = common.Big0
	}
	if args.Amount == nil || args.Amount.Sign() < 0 || args.Amount.Cmp(netProxied) == -1 {
		return nil, core.ErrVoteAmountTooLow
	}

	// Non Candidate, Check Amount greater than minimumVoteAmount
	if !state.IsCandidate(from) && args.Amount.Sign() == 1 && args.Amount.Cmp(minimumVoteAmount) == -1 {
		return nil, core.ErrVoteAmountTooLow
	}

	// Check Amount (Amount <= net proxied + balance + deposit)
	balance := state.GetBalance(from)
	deposit := state.GetDepositBalance(from)
	maximumAmount := new(big.Int).Add(new(big.Int).Add(balance, deposit), netProxied)
	if args.Amount.Cmp(maximumAmount) == 1 {
		return nil, core.ErrVoteAmountTooHight
	}

	// Check Signature of the PubKey matched against the Address
	if err := crypto.CheckConsensusPubKey(from, args.PubKey, args.Signature); err != nil {
		return nil, err
	}

	// Check Epoch Height
	ep, err := checkEpochInRevealVoteStage(bc)
	if err != nil {
		return nil, err
	}

	// Check Vote
	voteSet := ep.GetNextEpoch().GetEpochValidatorVoteSet()
	if voteSet == nil {
		return nil, errors.New(fmt.Sprintf("Can not found the vote for Address %v", from.String()))
	}

	vote, exist := voteSet.GetVoteByAddress(from)
	// Check Vote exist
	if !exist {
		return nil, errors.New(fmt.Sprintf("Can not found the vote for Address %v", from.String()))
	}

	if len(vote.VoteHash) == 0 {
		return nil, errors.New(fmt.Sprintf("Address %v doesn't has vote hash", from.String()))
	}

	// Check Vote Hash
	byte_data := [][]byte{
		from.Bytes(),
		args.PubKey,
		common.LeftPadBytes(args.Amount.Bytes(), 1),
		[]byte(args.Salt),
	}
	voteHash := intcrypto.Keccak256Hash(concatCopyPreAllocate(byte_data))
	fmt.Printf("tdm api voteHash %v\n", voteHash.String())
	if vote.VoteHash != voteHash {
		return nil, errors.New("your vote doesn't match your vote hash, please check your vote")
	}

	// Check Logic - Amount can't be 0 for new Validator
	if !ep.Validators.HasAddress(from.Bytes()) && args.Amount.Sign() == 0 {
		return nil, errors.New("invalid vote!!! new validator's vote amount must be greater than 0")
	}

	// Check Logic - SuperNode with remaining epoch can not decrease the stack
	if _, supernode := ep.Validators.GetByAddress(from.Bytes()); supernode != nil {
		if supernode.RemainingEpoch > 0 && args.Amount.Cmp(state.GetDepositBalance(from)) == -1 {
			return nil, core.ErrVoteAmountTooLow
		}
	}

	return &args, nil
}

// Common

func checkEpochInHashVoteStage(bc *core.BlockChain) error {
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}

	if ep == nil {
		return errors.New("epoch is nil, are you running on Tendermint Consensus Engine")
	}

	// Check Epoch in Hash Vote stage
	if ep.GetNextEpoch() == nil {
		return errors.New("next Epoch is nil, You can't vote the next epoch")
	}

	// Vote is valid between height 75% - 85%
	//height := bc.CurrentBlock().NumberU64()
	//if !ep.CheckInHashVoteStage(height) {
	//	return errors.New(fmt.Sprintf("you can't send the hash vote during this time, current height %v", height))
	//}
	return nil
}

func checkEpochInRevealVoteStage(bc *core.BlockChain) (*epoch.Epoch, error) {
	ep, err := getEpoch(bc)
	if err != nil {
		return nil, err
	}
	// Check Epoch in Reveal Vote stage
	if ep.GetNextEpoch() == nil {
		return nil, errors.New("next Epoch is nil, You can't vote the next epoch")
	}

	// Vote is valid between height 85% - 95%
	//height := bc.CurrentBlock().NumberU64()
	//if !ep.CheckInRevealVoteStage(height) {
	//	return nil, errors.New(fmt.Sprintf("you can't send the reveal vote during this time, current height %v", height))
	//}
	return ep, nil
}

func concatCopyPreAllocate(slices [][]byte) []byte {
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

func unForbidValidation(from common.Address, state *state.StateDB, bc *core.BlockChain) error {
	if !state.IsCandidate(from) {
		return core.ErrNotCandidate
	}

	ep, err := getEpoch(bc)
	if err != nil {
		return err
	}

	fromObj := state.GetOrNewStateObject(from)
	isForbidden := fromObj.IsForbidden()
	if !isForbidden {
		return fmt.Errorf("should not unforbid")
	}

	forbiddenDuration := ep.GetForbiddenDuration()
	forbiddenTime := fromObj.BlockTime()

	durationToNow := new(big.Int).Sub(big.NewInt(time.Now().Unix()), forbiddenTime)
	if durationToNow.Cmp(big.NewInt(int64(forbiddenDuration.Seconds()))) < 0 {
		return fmt.Errorf("time is too short to unforbid, forbidden duration %v, but duratrion to now %v", forbiddenDuration.Seconds(), durationToNow)
	}
	return nil
}

func getEpoch(bc *core.BlockChain) (*epoch.Epoch, error) {
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}

	if ep == nil {
		return nil, errors.New("epoch is nil, are you running on IPBFT Consensus Engine")
	}

	return ep, nil
}
