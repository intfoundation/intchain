package ipbft

import (
	"errors"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft/epoch"
	tdmTypes "github.com/intfoundation/intchain/consensus/ipbft/types"
	intCrypto "github.com/intfoundation/intchain/crypto"
	"math/big"
)

// API is a user facing RPC API of Tendermint
type API struct {
	chain      consensus.ChainReader
	tendermint *backend
}

// GetCurrentEpochNumber retrieves the current epoch number.
func (api *API) GetCurrentEpochNumber() (hexutil.Uint64, error) {
	return hexutil.Uint64(api.tendermint.core.consensusState.Epoch.Number), nil
}

// GetEpoch retrieves the Epoch Detail by Number
func (api *API) GetEpoch(num hexutil.Uint64) (*tdmTypes.EpochApi, error) {

	number := uint64(num)
	var resultEpoch *epoch.Epoch
	curEpoch := api.tendermint.core.consensusState.Epoch
	if number < 0 || number > curEpoch.Number {
		return nil, errors.New("epoch number out of range")
	}

	if number == curEpoch.Number {
		resultEpoch = curEpoch
	} else {
		resultEpoch = epoch.LoadOneEpoch(curEpoch.GetDB(), number, nil)
	}

	validators := make([]*tdmTypes.EpochValidator, len(resultEpoch.Validators.Validators))
	for i, val := range resultEpoch.Validators.Validators {
		validators[i] = &tdmTypes.EpochValidator{
			Address:        common.BytesToAddress(val.Address),
			PubKey:         val.PubKey.KeyString(),
			Amount:         (*hexutil.Big)(val.VotingPower),
			RemainingEpoch: hexutil.Uint64(val.RemainingEpoch),
		}
	}

	return &tdmTypes.EpochApi{
		Number:         hexutil.Uint64(resultEpoch.Number),
		RewardPerBlock: (*hexutil.Big)(resultEpoch.RewardPerBlock),
		StartBlock:     hexutil.Uint64(resultEpoch.StartBlock),
		EndBlock:       hexutil.Uint64(resultEpoch.EndBlock),
		StartTime:      resultEpoch.StartTime,
		EndTime:        resultEpoch.EndTime,
		Validators:     validators,
	}, nil
}

// GetEpochVote
func (api *API) GetNextEpochVote() (*tdmTypes.EpochVotesApi, error) {

	ep := api.tendermint.core.consensusState.Epoch
	if ep.GetNextEpoch() != nil {

		var votes []*epoch.EpochValidatorVote
		if ep.GetNextEpoch().GetEpochValidatorVoteSet() != nil {
			votes = ep.GetNextEpoch().GetEpochValidatorVoteSet().Votes
		}
		votesApi := make([]*tdmTypes.EpochValidatorVoteApi, 0, len(votes))
		for _, v := range votes {
			var pkstring string
			if v.PubKey != nil {
				pkstring = v.PubKey.KeyString()
			}

			votesApi = append(votesApi, &tdmTypes.EpochValidatorVoteApi{
				EpochValidator: tdmTypes.EpochValidator{
					Address: v.Address,
					PubKey:  pkstring,
					Amount:  (*hexutil.Big)(v.Amount),
				},
				Salt:     v.Salt,
				VoteHash: v.VoteHash,
				TxHash:   v.TxHash,
			})
		}

		return &tdmTypes.EpochVotesApi{
			EpochNumber: hexutil.Uint64(ep.GetNextEpoch().Number),
			StartBlock:  hexutil.Uint64(ep.GetNextEpoch().StartBlock),
			EndBlock:    hexutil.Uint64(ep.GetNextEpoch().EndBlock),
			Votes:       votesApi,
		}, nil
	}
	return nil, errors.New("next epoch has not been proposed")
}

func (api *API) GetNextEpochValidators() ([]*tdmTypes.EpochValidator, error) {

	//height := api.chain.CurrentBlock().NumberU64()

	ep := api.tendermint.core.consensusState.Epoch
	nextEp := ep.GetNextEpoch()
	if nextEp == nil {
		return nil, errors.New("voting for next epoch has not started yet")
	} else {
		state, err := api.chain.State()
		if err != nil {
			return nil, err
		}

		nextValidators := ep.Validators.Copy()
		err = epoch.DryRunUpdateEpochValidatorSet(state, nextValidators, nextEp.GetEpochValidatorVoteSet())
		if err != nil {
			return nil, err
		}

		validators := make([]*tdmTypes.EpochValidator, 0, len(nextValidators.Validators))
		for _, val := range nextValidators.Validators {
			var pkstring string
			if val.PubKey != nil {
				pkstring = val.PubKey.KeyString()
			}
			validators = append(validators, &tdmTypes.EpochValidator{
				Address:        common.BytesToAddress(val.Address),
				PubKey:         pkstring,
				Amount:         (*hexutil.Big)(val.VotingPower),
				RemainingEpoch: hexutil.Uint64(val.RemainingEpoch),
			})
		}

		return validators, nil
	}
}

// CreateValidator no longer support
//func (api *API) CreateValidator(from common.Address) (*tdmTypes.PrivValidator, error) {
//	validator := tdmTypes.GenPrivValidatorKey(from)
//	return validator, nil
//}

// decode extra data
func (api *API) DecodeExtraData(extra string) (extraApi *tdmTypes.TendermintExtraApi, err error) {
	tdmExtra, err := tdmTypes.DecodeExtraData(extra)
	if err != nil {
		return nil, err
	}
	extraApi = &tdmTypes.TendermintExtraApi{
		ChainID:         tdmExtra.ChainID,
		Height:          hexutil.Uint64(tdmExtra.Height),
		Time:            tdmExtra.Time,
		NeedToSave:      tdmExtra.NeedToSave,
		NeedToBroadcast: tdmExtra.NeedToBroadcast,
		EpochNumber:     hexutil.Uint64(tdmExtra.EpochNumber),
		SeenCommitHash:  hexutil.Encode(tdmExtra.SeenCommitHash),
		ValidatorsHash:  hexutil.Encode(tdmExtra.ValidatorsHash),
		SeenCommit: &tdmTypes.CommitApi{
			BlockID: tdmTypes.BlockIDApi{
				Hash: hexutil.Encode(tdmExtra.SeenCommit.BlockID.Hash),
				PartsHeader: tdmTypes.PartSetHeaderApi{
					Total: hexutil.Uint64(tdmExtra.SeenCommit.BlockID.PartsHeader.Total),
					Hash:  hexutil.Encode(tdmExtra.SeenCommit.BlockID.PartsHeader.Hash),
				},
			},
			Height:   hexutil.Uint64(tdmExtra.SeenCommit.Height),
			Round:    tdmExtra.SeenCommit.Round,
			SignAggr: tdmExtra.SeenCommit.SignAggr,
			BitArray: tdmExtra.SeenCommit.BitArray,
		},
		EpochBytes: tdmExtra.EpochBytes,
	}
	return extraApi, nil
}

// get consensus publickey of the block
func (api *API) GetConsensusPublicKey(extra string) ([]string, error) {
	tdmExtra, err := tdmTypes.DecodeExtraData(extra)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("GetConsensusPublicKey tdmExtra %v\n", tdmExtra)
	number := uint64(tdmExtra.EpochNumber)
	var resultEpoch *epoch.Epoch
	curEpoch := api.tendermint.core.consensusState.Epoch
	if number < 0 || number > curEpoch.Number {
		return nil, errors.New("epoch number out of range")
	}

	if number == curEpoch.Number {
		resultEpoch = curEpoch
	} else {
		resultEpoch = epoch.LoadOneEpoch(curEpoch.GetDB(), number, nil)
	}

	//fmt.Printf("GetConsensusPublicKey result epoch %v\n", resultEpoch)
	validatorSet := resultEpoch.Validators
	//fmt.Printf("GetConsensusPublicKey validatorset %v\n", validatorSet)

	aggr, err := validatorSet.GetAggrPubKeyAndAddress(tdmExtra.SeenCommit.BitArray)
	if err != nil {
		return nil, err
	}

	var pubkeys []string
	if len(aggr.PublicKeys) > 0 {
		for _, v := range aggr.PublicKeys {
			if v != "" {
				pubkeys = append(pubkeys, v)
			}
		}
	}

	return pubkeys, nil
}

func (api *API) GetVoteHash(from common.Address, pubkey crypto.BLSPubKey, amount *hexutil.Big, salt string) common.Hash {
	byteData := [][]byte{
		from.Bytes(),
		pubkey.Bytes(),
		(*big.Int)(amount).Bytes(),
		[]byte(salt),
	}
	return intCrypto.Keccak256Hash(ConcatCopyPreAllocate(byteData))
}

//func (api *API) GetValidatorStatus(from common.Address) (*tdmTypes.ValidatorStatus, error) {
//	state, err := api.chain.State()
//	if state == nil || err != nil {
//		return nil, err
//	}
//	status := &tdmTypes.ValidatorStatus{
//		IsForbidden: state.GetOrNewStateObject(from).IsForbidden(),
//	}
//
//	return status, nil
//}

//func (api *API) GetCandidateList() (*tdmTypes.CandidateApi, error) {
//	state, err := api.chain.State()
//
//	if state == nil || err != nil {
//		return nil, err
//	}
//
//	candidateList := make([]string, 0)
//	candidateSet := state.GetCandidateSet()
//	fmt.Printf("candidate set %v", candidateSet)
//	for addr := range candidateSet {
//		candidateList = append(candidateList, addr)
//	}
//
//	candidates := &tdmTypes.CandidateApi{
//		CandidateList: candidateList,
//	}
//
//	return candidates, nil
//}
//
//func (api *API) GetForbiddenList() (*tdmTypes.ForbiddenApi, error) {
//	state, err := api.chain.State()
//
//	if state == nil || err != nil {
//		return nil, err
//	}
//
//	forbiddenList := make([]string, 0)
//	forbiddenSet := state.GetForbiddenSet()
//	fmt.Printf("forbidden set %v", forbiddenSet)
//	for addr := range forbiddenSet {
//		forbiddenList = append(forbiddenList, addr)
//	}
//
//	forbiddenAddresses := &tdmTypes.ForbiddenApi{
//		ForbiddenList: forbiddenList,
//	}
//
//	return forbiddenAddresses, nil
//}
