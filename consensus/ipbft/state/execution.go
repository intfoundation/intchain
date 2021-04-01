package state

import (
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/consensus"
	ep "github.com/intfoundation/intchain/consensus/ipbft/epoch"
	"github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/core"
	ethTypes "github.com/intfoundation/intchain/core/types"
)

//--------------------------------------------------

// return a bit array of validators that signed the last commit
// NOTE: assumes commits have already been authenticated
/*
func commitBitArrayFromBlock(block *types.TdmBlock) *BitArray {

	signed := NewBitArray(uint64(len(block.TdmExtra.SeenCommit.Precommits)))
	for i, precommit := range block.TdmExtra.SeenCommit.Precommits {
		if precommit != nil {
			signed.SetIndex(uint64(i), true) // val_.LastCommitHeight = block.Height - 1
		}
	}
	return signed
}*/

//-----------------------------------------------------
// Validate block

func (s *State) ValidateBlock(block *types.TdmBlock) error {
	return s.validateBlock(block)
}

//Very current block
func (s *State) validateBlock(block *types.TdmBlock) error {
	// Basic block validation.
	err := block.ValidateBasic(s.TdmExtra)
	if err != nil {
		return err
	}

	// Validate block SeenCommit.
	epoch := s.Epoch.GetEpochByBlockNumber(block.TdmExtra.Height)
	if epoch == nil || epoch.Validators == nil {
		return errors.New("no epoch for current block height")
	}

	valSet := epoch.Validators
	err = valSet.VerifyCommit(block.TdmExtra.ChainID, block.TdmExtra.Height,
		block.TdmExtra.SeenCommit)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

func init() {
	core.RegisterInsertBlockCb("UpdateLocalEpoch", updateLocalEpoch)
	core.RegisterInsertBlockCb("AutoStartMining", autoStartMining)
}

func updateLocalEpoch(bc *core.BlockChain, block *ethTypes.Block) {
	if block.NumberU64() == 0 {
		return
	}

	tdmExtra, _ := types.ExtractTendermintExtra(block.Header())
	//here handles the proposed next epoch
	epochInBlock := ep.FromBytes(tdmExtra.EpochBytes)

	eng := bc.Engine().(consensus.IPBFT)
	currentEpoch := eng.GetEpoch()

	if epochInBlock != nil {
		fmt.Printf("update local epoch: epochInBlock %v\n", epochInBlock)
		fmt.Printf("update local epoch: tdmExtra %v\n", tdmExtra)
		if epochInBlock.Number == currentEpoch.Number+1 {
			//fmt.Printf("update local epoch 1\n")
			// Save the next epoch
			if block.NumberU64() == currentEpoch.StartBlock+2 {
				//fmt.Printf("update local epoch block number %v, current epoch start block %v\n", block.NumberU64(), currentEpoch.StartBlock)
				// Propose next epoch
				//epochInBlock.SetEpochValidatorVoteSet(ep.NewEpochValidatorVoteSet())
				epochInBlock.Status = ep.EPOCH_VOTED_NOT_SAVED
				epochInBlock.SetRewardScheme(currentEpoch.GetRewardScheme())
				currentEpoch.SetNextEpoch(epochInBlock)
			} else if block.NumberU64() == currentEpoch.EndBlock {
				//fmt.Printf("update local epoch 2\n")
				// Finalize next epoch
				// Validator set in next epoch will not finalize and send to main chain
				nextEp := currentEpoch.GetNextEpoch()
				nextEp.Validators = epochInBlock.Validators
				nextEp.Status = ep.EPOCH_VOTED_NOT_SAVED
			}
			currentEpoch.Save()
		} else if epochInBlock.Number == currentEpoch.Number {
			fmt.Printf("update local epoch: epochInBlock.Number %v, currentEpoch.Number %v, epochInBlock.StartTime %v\n", epochInBlock.Number, currentEpoch.Number, epochInBlock.StartTime)
			// Update the current epoch Start Time from proposer
			currentEpoch.StartTime = epochInBlock.StartTime
			currentEpoch.Save()

			// Update the previous epoch End Time
			if currentEpoch.Number > 0 {
				ep.UpdateEpochEndTime(currentEpoch.GetDB(), currentEpoch.Number-1, epochInBlock.StartTime)
			}
		}
	}
}

func autoStartMining(bc *core.BlockChain, block *ethTypes.Block) {
	eng := bc.Engine().(consensus.IPBFT)
	currentEpoch := eng.GetEpoch()

	// At one block before epoch end block, we should able to calculate the new validator
	if block.NumberU64() == currentEpoch.EndBlock-1 {
		fmt.Printf("auto start mining first %v\n", block.Number())
		// Re-Calculate the next epoch validators
		nextEp := currentEpoch.GetNextEpoch()
		state, _ := bc.State()
		nextValidators := currentEpoch.Validators.Copy()
		dryrunErr := ep.DryRunUpdateEpochValidatorSet(state, nextValidators, nextEp.GetEpochValidatorVoteSet())
		if dryrunErr != nil {
			panic("can not update the validator set base on the vote, error: " + dryrunErr.Error())
		}
		nextEp.Validators = nextValidators

		if nextValidators.HasAddress(eng.PrivateValidator().Bytes()) && !eng.IsStarted() {
			fmt.Printf("auto start mining first, post start mining event")
			bc.PostChainEvents([]interface{}{core.StartMiningEvent{}}, nil)
		}
	}
}
