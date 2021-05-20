package epoch

import (
	"errors"
	"fmt"
	dbm "github.com/intfoundation/go-db"
	"github.com/intfoundation/go-wire"
	"github.com/intfoundation/intchain/common"
	tmTypes "github.com/intfoundation/intchain/consensus/ipbft/types"
	"github.com/intfoundation/intchain/core/state"
	"github.com/intfoundation/intchain/log"
	//"math"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"
)

var NextEpochNotExist = errors.New("next epoch parameters do not exist, fatal error")
var NextEpochNotEXPECTED = errors.New("next epoch parameters are not excepted, fatal error")

var ForbiddenEpoch = big.NewInt(2) // forbid 2 epoch

const (
	EPOCH_NOT_EXIST          = iota // value --> 0
	EPOCH_PROPOSED_NOT_VOTED        // value --> 1
	EPOCH_VOTED_NOT_SAVED           // value --> 2
	EPOCH_SAVED                     // value --> 3

	MinimumValidatorsSize = 13
	MaximumValidatorsSize = 25

	epochKey       = "Epoch:%v"
	latestEpochKey = "LatestEpoch"

	ForbiddenDuration = 24 * time.Hour
)

type Epoch struct {
	mtx sync.Mutex
	db  dbm.DB

	Number         uint64
	RewardPerBlock *big.Int
	StartBlock     uint64
	EndBlock       uint64
	StartTime      time.Time
	EndTime        time.Time //not accurate for current epoch
	BlockGenerated int       //agreed in which block
	Status         int       //checked if this epoch has been saved
	Validators     *tmTypes.ValidatorSet

	// The VoteSet will be used just before Epoch Start
	validatorVoteSet *EpochValidatorVoteSet // VoteSet store with key prefix EpochValidatorVote_
	rs               *RewardScheme          // RewardScheme store with key REWARDSCHEME
	previousEpoch    *Epoch
	nextEpoch        *Epoch

	logger log.Logger
}

func calcEpochKeyWithHeight(number uint64) []byte {
	return []byte(fmt.Sprintf(epochKey, number))
}

// InitEpoch either initial the Epoch from DB or from genesis file
func InitEpoch(db dbm.DB, genDoc *tmTypes.GenesisDoc, logger log.Logger) *Epoch {

	epochNumber := db.Get([]byte(latestEpochKey))
	if epochNumber == nil {
		// Read Epoch from Genesis
		rewardScheme := MakeRewardScheme(db, &genDoc.RewardScheme)
		rewardScheme.Save()

		ep := MakeOneEpoch(db, &genDoc.CurrentEpoch, logger)
		ep.Save()

		ep.SetRewardScheme(rewardScheme)
		return ep
	} else {
		// Load Epoch from DB
		epNo, _ := strconv.ParseUint(string(epochNumber), 10, 64)
		return LoadOneEpoch(db, epNo, logger)
	}
}

// Load Full Epoch By EpochNumber (Epoch data, Reward Scheme, ValidatorVote, Previous Epoch, Next Epoch)
func LoadOneEpoch(db dbm.DB, epochNumber uint64, logger log.Logger) *Epoch {
	// Load Epoch Data from DB
	epoch := loadOneEpoch(db, epochNumber, logger)
	// Set Reward Scheme
	rewardscheme := LoadRewardScheme(db)
	epoch.rs = rewardscheme
	// Set Validator VoteSet if has
	epoch.validatorVoteSet = LoadEpochVoteSet(db, epochNumber)
	// Set Previous Epoch
	if epochNumber > 0 {
		epoch.previousEpoch = loadOneEpoch(db, epochNumber-1, logger)
		if epoch.previousEpoch != nil {
			epoch.previousEpoch.rs = rewardscheme
		}
	}
	// Set Next Epoch
	epoch.nextEpoch = loadOneEpoch(db, epochNumber+1, logger)
	if epoch.nextEpoch != nil {
		epoch.nextEpoch.rs = rewardscheme
		// Set ValidatorVoteSet
		epoch.nextEpoch.validatorVoteSet = LoadEpochVoteSet(db, epochNumber+1)
	}

	return epoch
}

func loadOneEpoch(db dbm.DB, epochNumber uint64, logger log.Logger) *Epoch {

	buf := db.Get(calcEpochKeyWithHeight(epochNumber))
	ep := FromBytes(buf)
	if ep != nil {
		ep.db = db
		ep.logger = logger
	}
	return ep
}

// Convert from OneEpochDoc (Json) to Epoch
func MakeOneEpoch(db dbm.DB, oneEpoch *tmTypes.OneEpochDoc, logger log.Logger) *Epoch {

	//fmt.Printf("MakeOneEpoch onEpoch=%v\n", oneEpoch)
	//fmt.Printf("MakeOneEpoch validators=%v\n", oneEpoch.Validators)
	validators := make([]*tmTypes.Validator, len(oneEpoch.Validators))
	for i, val := range oneEpoch.Validators {
		// Make validator
		validators[i] = &tmTypes.Validator{
			Address:        val.EthAccount.Bytes(),
			PubKey:         val.PubKey,
			VotingPower:    val.Amount,
			RemainingEpoch: val.RemainingEpoch,
		}
	}

	te := &Epoch{
		db: db,

		Number:         oneEpoch.Number,
		RewardPerBlock: oneEpoch.RewardPerBlock,
		StartBlock:     oneEpoch.StartBlock,
		EndBlock:       oneEpoch.EndBlock,
		StartTime:      time.Now(),
		EndTime:        time.Unix(0, 0), //not accurate for current epoch
		Status:         oneEpoch.Status,
		Validators:     tmTypes.NewValidatorSet(validators),

		logger: logger,
	}

	return te
}

func (epoch *Epoch) GetDB() dbm.DB {
	return epoch.db
}

func (epoch *Epoch) GetEpochValidatorVoteSet() *EpochValidatorVoteSet {
	//try reload validatorVoteSet
	if epoch.validatorVoteSet == nil {
		epoch.validatorVoteSet = LoadEpochVoteSet(epoch.db, epoch.Number)
	}
	return epoch.validatorVoteSet
}

func (epoch *Epoch) GetRewardScheme() *RewardScheme {
	return epoch.rs
}

func (epoch *Epoch) SetRewardScheme(rs *RewardScheme) {
	epoch.rs = rs
}

// Save the Epoch to Level DB
func (epoch *Epoch) Save() {
	epoch.mtx.Lock()
	defer epoch.mtx.Unlock()
	epoch.db.SetSync(calcEpochKeyWithHeight(epoch.Number), epoch.Bytes())
	epoch.db.SetSync([]byte(latestEpochKey), []byte(strconv.FormatUint(epoch.Number, 10)))

	if epoch.nextEpoch != nil && epoch.nextEpoch.Status == EPOCH_VOTED_NOT_SAVED {
		epoch.nextEpoch.Status = EPOCH_SAVED
		// Save the next epoch
		epoch.db.SetSync(calcEpochKeyWithHeight(epoch.nextEpoch.Number), epoch.nextEpoch.Bytes())
	}

	// TODO whether save next epoch validator vote set
	//if epoch.nextEpoch != nil && epoch.nextEpoch.validatorVoteSet != nil {
	//	// Save the next epoch vote set
	//	SaveEpochVoteSet(epoch.db, epoch.nextEpoch.Number, epoch.nextEpoch.validatorVoteSet)
	//}
}

func FromBytes(buf []byte) *Epoch {

	if len(buf) == 0 {
		return nil
	} else {
		ep := &Epoch{}
		err := wire.ReadBinaryBytes(buf, ep)
		if err != nil {
			log.Errorf("Load Epoch from Bytes Failed, error: %v", err)
			return nil
		}
		return ep
	}
}

func (epoch *Epoch) Bytes() []byte {
	return wire.BinaryBytes(*epoch)
}

func (epoch *Epoch) ValidateNextEpoch(next *Epoch, lastHeight uint64, lastBlockTime time.Time) error {

	myNextEpoch := epoch.ProposeNextEpoch(lastHeight, lastBlockTime)

	if !myNextEpoch.Equals(next, false) {
		log.Warnf("next epoch parameters are not expected, epoch propose next epoch: %v, next %v", myNextEpoch.String(), next.String())
		return NextEpochNotEXPECTED
	}

	return nil
}

//check if need propose next epoch
func (epoch *Epoch) ShouldProposeNextEpoch(curBlockHeight uint64) bool {
	// If next epoch already proposed, then no need propose again
	//fmt.Printf("should propose next epoch %v\n", epoch.nextEpoch)
	if epoch.nextEpoch != nil {
		return false
	}

	// current block height bigger than epoch start block + 1 and not equal to epoch end block
	shouldPropose := curBlockHeight > (epoch.StartBlock+1) && curBlockHeight != epoch.EndBlock
	return shouldPropose
}

func (epoch *Epoch) ProposeNextEpoch(lastBlockHeight uint64, lastBlockTime time.Time) *Epoch {

	if epoch != nil {

		rewardPerBlock, blocks := epoch.estimateForNextEpoch(lastBlockHeight, lastBlockTime)

		next := &Epoch{
			mtx: epoch.mtx,
			db:  epoch.db,

			Number:         epoch.Number + 1,
			RewardPerBlock: rewardPerBlock,
			StartBlock:     epoch.EndBlock + 1,
			EndBlock:       epoch.EndBlock + blocks,
			BlockGenerated: 0,
			Status:         EPOCH_PROPOSED_NOT_VOTED,
			Validators:     epoch.Validators.Copy(), // Old Validators

			logger: epoch.logger,
		}

		return next
	}
	return nil
}

func (epoch *Epoch) GetNextEpoch() *Epoch {
	if epoch.nextEpoch == nil {
		epoch.nextEpoch = loadOneEpoch(epoch.db, epoch.Number+1, epoch.logger)
		if epoch.nextEpoch != nil {
			epoch.nextEpoch.rs = epoch.rs
			// Set ValidatorVoteSet
			epoch.nextEpoch.validatorVoteSet = LoadEpochVoteSet(epoch.db, epoch.Number+1)
		}
	}
	return epoch.nextEpoch
}

func (epoch *Epoch) SetNextEpoch(next *Epoch) {
	if next != nil {
		next.db = epoch.db
		next.rs = epoch.rs
		next.logger = epoch.logger
	}
	epoch.nextEpoch = next
}

func (epoch *Epoch) GetPreviousEpoch() *Epoch {
	return epoch.previousEpoch
}

func (epoch *Epoch) ShouldEnterNewEpoch(height uint64, state *state.StateDB) (bool, *tmTypes.ValidatorSet, error) {

	if height == epoch.EndBlock {
		epoch.nextEpoch = epoch.GetNextEpoch()
		if epoch.nextEpoch != nil {

			// Step 1: Refund the Delegate (subtract the pending refund / deposit proxied amount)
			for refundAddress := range state.GetDelegateAddressRefundSet() {
				state.ForEachProxied(refundAddress, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
					if pendingRefundBalance.Sign() > 0 {
						// Refund Pending Refund
						state.SubDepositProxiedBalanceByUser(refundAddress, key, pendingRefundBalance)
						state.SubPendingRefundBalanceByUser(refundAddress, key, pendingRefundBalance)
						state.SubDelegateBalance(key, pendingRefundBalance)
						state.AddBalance(key, pendingRefundBalance)
					}
					return true
				})
				// reset commission = 0 if not candidate
				if !state.IsCandidate(refundAddress) {
					state.ClearCommission(refundAddress)
				}
			}
			state.ClearDelegateRefundSet()

			// Step 2: Sort the Validators and potential Validators (with success vote) base on deposit amount + deposit proxied amount
			// Step 2.1: Update deposit amount base on the vote (Add/Substract deposit amount base on vote)
			// Step 2.2: Add candidate to next epoch vote set
			// Step 2.3: Sort the address with deposit + deposit proxied amount
			var (
				refunds []*tmTypes.RefundValidatorAmount
			)

			newValidators := epoch.Validators.Copy()
			// Invoke the get next epoch method to avoid next epoch vote set is nil
			//nextEpochVoteSet := epoch.GetNextEpoch().GetEpochValidatorVoteSet().Copy() // copy vote set
			//candidateList := state.GetCandidateSet()

			for _, v := range newValidators.Validators {
				vAddr := common.BytesToAddress(v.Address)
				//if !state.GetForbidden(vAddr) {
				//epoch.logger.Debugf("Should enter new epoch, validator %v is not forbidden", vAddr)
				totalProxiedBalance := new(big.Int).Add(state.GetTotalProxiedBalance(vAddr), state.GetTotalDepositProxiedBalance(vAddr))
				// Voting Power = Proxied amount + Deposit amount
				newVotingPower := new(big.Int).Add(totalProxiedBalance, state.GetDepositBalance(vAddr))
				if newVotingPower.Sign() == 0 {
					newValidators.Remove(v.Address)
				} else {
					v.VotingPower = newVotingPower
				}
				//} else {
				//	epoch.logger.Debugf("Should enter new epoch, validator %v is forbidden", vAddr)
				//	// if forbidden then remove from the validator set and candidate list
				//	newValidators.Remove(v.Address)
				//	delete(candidateList, vAddr)
				//
				//	// if forbidden epoch bigger than 0, subtract 1 epoch
				//	forbiddenEpoch := state.GetForbiddenTime(vAddr)
				//	if forbiddenEpoch.Cmp(common.Big0) == 1 {
				//		forbiddenEpoch.Sub(forbiddenEpoch, common.Big1)
				//		state.SetForbiddenTime(vAddr, forbiddenEpoch)
				//		epoch.logger.Debugf("Should enter new epoch 1, left forbidden epoch is %v", forbiddenEpoch)
				//	}
				//
				//	refunds = append(refunds, &tmTypes.RefundValidatorAmount{Address: vAddr, Amount: v.VotingPower, Voteout: true})
				//}
			}

			//if nextEpochVoteSet == nil {
			//	nextEpochVoteSet = NewEpochValidatorVoteSet()
			//	epoch.logger.Debugf("Should enter new epoch, next epoch vote set is nil, %v", nextEpochVoteSet)
			//}

			// if has candidate and next epoch vote set not nil, add them to next epoch vote set
			//if len(candidateList) > 0 {
			//	for addr := range candidateList {
			//		if state.GetForbidden(addr) {
			//			// first, delete from the candidate list
			//			delete(candidateList, addr)
			//
			//			// if forbidden epoch bigger than 0, subtract 1 epoch
			//			forbiddenEpoch := state.GetForbiddenTime(addr)
			//			if forbiddenEpoch.Cmp(common.Big0) == 1 {
			//				forbiddenEpoch.Sub(forbiddenEpoch, common.Big1)
			//				state.SetForbiddenTime(addr, forbiddenEpoch)
			//				//epoch.logger.Debugf("Should enter new epoch 2, left forbidden epoch is %v\n", forbiddenEpoch)
			//			}
			//		}
			//	}
			//
			//	epoch.logger.Debugf("Add candidate to next epoch vote set before, candidate: %v", candidateList)
			//
			//	for _, v := range newValidators.Validators {
			//		vAddr := common.BytesToAddress(v.Address)
			//		delete(candidateList, vAddr)
			//	}
			//
			//	for _, v := range nextEpochVoteSet.Votes {
			//		// first, delete from the candidate list
			//		delete(candidateList, v.Address)
			//	}
			//
			//	epoch.logger.Debugf("Add candidate to next epoch vote set after, candidate: %v", candidateList)
			//
			//	var voteArr []*EpochValidatorVote
			//	for addr := range candidateList {
			//		if state.IsCandidate(addr) {
			//			// calculate the net proxied balance of this candidate
			//			proxiedBalance := state.GetTotalProxiedBalance(addr)
			//			// TODO if need add the deposit proxied balance
			//			depositProxiedBalance := state.GetTotalDepositProxiedBalance(addr)
			//			// TODO if need subtraction the pending refund balance
			//			pendingRefundBalance := state.GetTotalPendingRefundBalance(addr)
			//			netProxied := new(big.Int).Sub(new(big.Int).Add(proxiedBalance, depositProxiedBalance), pendingRefundBalance)
			//
			//			if netProxied.Sign() == -1 {
			//				continue
			//			}
			//
			//			// TODO whether need move the delegate amount now
			//			// Move delegate amount first if Candidate
			//			state.ForEachProxied(addr, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
			//				// Move Proxied Amount to Deposit Proxied Amount
			//				state.SubProxiedBalanceByUser(addr, key, proxiedBalance)
			//				state.AddDepositProxiedBalanceByUser(addr, key, proxiedBalance)
			//				return true
			//			})
			//
			//			pubkey := state.GetPubkey(addr)
			//			pubkeyBytes := common.FromHex(pubkey)
			//			if pubkey == "" || len(pubkeyBytes) != 128 {
			//				continue
			//			}
			//			var blsPK goCrypto.BLSPubKey
			//			copy(blsPK[:], pubkeyBytes)
			//
			//			vote := &EpochValidatorVote{
			//				Address: addr,
			//				Amount:  netProxied,
			//				PubKey:  blsPK,
			//				Salt:    "intchain",
			//				TxHash:  common.Hash{},
			//			}
			//			voteArr = append(voteArr, vote)
			//			fmt.Printf("vote %v\n", vote)
			//			//nextEpochVoteSet.StoreVote(vote)
			//		}
			//	}
			//
			//	// Sort the vote by amount and address
			//	sort.Slice(voteArr, func(i, j int) bool {
			//		if voteArr[i].Amount.Cmp(voteArr[j].Amount) == 0 {
			//			return compareAddress(voteArr[i].Address[:], voteArr[j].Address[:])
			//		} else {
			//			return voteArr[i].Amount.Cmp(voteArr[j].Amount) == 1
			//		}
			//	})
			//
			//	// Store the vote
			//	for i := range voteArr {
			//		epoch.logger.Debugf("address:%v, amount: %v\n", voteArr[i].Address, voteArr[i].Amount)
			//		nextEpochVoteSet.StoreVote(voteArr[i])
			//	}
			//}

			// Update Validators with vote
			refundsUpdate, err := updateEpochValidatorSet(newValidators, epoch.nextEpoch.validatorVoteSet)
			//refundsUpdate, err := updateEpochValidatorSet(newValidators, nextEpochVoteSet)
			if err != nil {
				epoch.logger.Warn("Error changing validator set", "error", err)
				return false, nil, err
			}
			refunds = append(refunds, refundsUpdate...)

			// Now newValidators become a real new Validators
			// Step 3: Special Case: For the existing Validator + Candidate + no vote, Move proxied amount to deposit proxied amount  (proxied amount -> deposit proxied amount)
			for _, v := range newValidators.Validators {
				vAddr := common.BytesToAddress(v.Address)
				if state.IsCandidate(vAddr) && state.GetTotalProxiedBalance(vAddr).Sign() > 0 {
					state.ForEachProxied(vAddr, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
						if proxiedBalance.Sign() > 0 {
							// Deposit the proxied amount
							state.SubProxiedBalanceByUser(vAddr, key, proxiedBalance)
							state.AddDepositProxiedBalanceByUser(vAddr, key, proxiedBalance)
						}
						return true
					})
				}
			}

			// Step 4: For vote out Address, refund deposit (deposit amount -> balance, deposit proxied amount -> proxied amount)
			for _, r := range refunds {
				if !r.Voteout {
					// Normal Refund, refund the deposit back to the self balance
					state.SubDepositBalance(r.Address, r.Amount)
					state.AddBalance(r.Address, r.Amount)
				} else {
					// Voteout Refund, refund the deposit both to self and proxied (if available)
					if state.IsCandidate(r.Address) {
						state.ForEachProxied(r.Address, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
							if depositProxiedBalance.Sign() > 0 {
								state.SubDepositProxiedBalanceByUser(r.Address, key, depositProxiedBalance)
								state.AddProxiedBalanceByUser(r.Address, key, depositProxiedBalance)
							}
							return true
						})
					}
					// Refund all the self deposit balance
					depositBalance := state.GetDepositBalance(r.Address)
					state.SubDepositBalance(r.Address, depositBalance)
					state.AddBalance(r.Address, depositBalance)
				}
			}

			return true, newValidators, nil
		} else {
			return false, nil, NextEpochNotExist
		}
	}
	return false, nil, nil
}

func compareAddress(addrA, addrB []byte) bool {
	if addrA[0] == addrB[0] {
		return compareAddress(addrA[1:], addrB[1:])
	} else {
		return addrA[0] > addrB[0]
	}
}

// Move to New Epoch
func (epoch *Epoch) EnterNewEpoch(newValidators *tmTypes.ValidatorSet) (*Epoch, error) {
	if epoch.nextEpoch != nil {
		now := time.Now()

		// Set the End Time for current Epoch and Save it
		epoch.EndTime = now
		epoch.Save()
		// Old Epoch Ended
		epoch.logger.Infof("Epoch %v reach to his end", epoch.Number)

		// Now move to Next Epoch
		nextEpoch := epoch.nextEpoch
		// Store the Previous Epoch Validators only
		nextEpoch.previousEpoch = &Epoch{Validators: epoch.Validators}
		// Store the Previous Epoch all
		//nextEpoch.previousEpoch = epoch.Copy() // if directly use epoch, it will falat error stack overflow (goroutine stack exceeds 1000000000-byte limit)

		nextEpoch.StartTime = now
		nextEpoch.Validators = newValidators

		nextEpoch.nextEpoch = nil //suppose we will not generate a more epoch after next-epoch
		nextEpoch.Save()
		epoch.logger.Infof("Enter into New Epoch %v", nextEpoch)
		return nextEpoch, nil
	} else {
		return nil, NextEpochNotExist
	}
}

// DryRunUpdateEpochValidatorSet Re-calculate the New Validator Set base on the current state db and vote set
func DryRunUpdateEpochValidatorSet(state *state.StateDB, validators *tmTypes.ValidatorSet, voteSet *EpochValidatorVoteSet) error {

	for _, v := range validators.Validators {
		vAddr := common.BytesToAddress(v.Address)

		// Deposit Proxied + Proxied - Pending Refund
		totalProxiedBalance := new(big.Int).Add(state.GetTotalProxiedBalance(vAddr), state.GetTotalDepositProxiedBalance(vAddr))
		totalProxiedBalance.Sub(totalProxiedBalance, state.GetTotalPendingRefundBalance(vAddr))

		// Voting Power = Delegated amount + Deposit amount
		newVotingPower := new(big.Int).Add(totalProxiedBalance, state.GetDepositBalance(vAddr))
		if newVotingPower.Sign() == 0 {
			validators.Remove(v.Address)
		} else {
			v.VotingPower = newVotingPower
		}
	}

	_, err := updateEpochValidatorSet(validators, voteSet)
	return err
}

// updateEpochValidatorSet Update the Current Epoch Validator by vote
//
func updateEpochValidatorSet(validators *tmTypes.ValidatorSet, voteSet *EpochValidatorVoteSet) ([]*tmTypes.RefundValidatorAmount, error) {

	// Refund List will be vaildators contain from Vote (exit validator or less amount than previous amount) and Knockout after sort by amount
	var refund []*tmTypes.RefundValidatorAmount
	oldValSize, newValSize := validators.Size(), 0

	// Process the Vote if vote set not empty
	if !voteSet.IsEmpty() {
		// Process the Votes and merge into the Validator Set
		for _, v := range voteSet.Votes {
			// If vote not reveal, bypass this vote
			if v.Amount == nil || v.Salt == "" || v.PubKey == nil {
				continue
			}

			_, validator := validators.GetByAddress(v.Address[:])
			if validator == nil {
				// Add the new validator
				added := validators.Add(tmTypes.NewValidator(v.Address[:], v.PubKey, v.Amount))
				if !added {
					return nil, fmt.Errorf("Failed to add new validator %v with voting power %d", v.Address, v.Amount)
				}
				newValSize++
			} else if v.Amount.Sign() == 0 {
				refund = append(refund, &tmTypes.RefundValidatorAmount{Address: v.Address, Amount: validator.VotingPower, Voteout: false})
				// Remove the Validator
				_, removed := validators.Remove(validator.Address)
				if !removed {
					return nil, fmt.Errorf("Failed to remove validator %v", validator.Address)
				}
			} else {
				//refund if new amount less than the voting power
				if v.Amount.Cmp(validator.VotingPower) == -1 {
					refundAmount := new(big.Int).Sub(validator.VotingPower, v.Amount)
					refund = append(refund, &tmTypes.RefundValidatorAmount{Address: v.Address, Amount: refundAmount, Voteout: false})
				}

				// Update the Validator Amount
				validator.VotingPower = v.Amount
				updated := validators.Update(validator)
				if !updated {
					return nil, fmt.Errorf("Failed to update validator %v with voting power %d", validator.Address, v.Amount)
				}
			}
		}
	}

	// Determine the Validator Size
	valSize := oldValSize + newValSize

	if valSize > MaximumValidatorsSize {
		valSize = MaximumValidatorsSize
	} else if valSize < MinimumValidatorsSize {
		valSize = MinimumValidatorsSize
	}

	// Subtract the remaining epoch value
	for _, v := range validators.Validators {
		if v.RemainingEpoch > 0 {
			v.RemainingEpoch--
		}
	}

	// If actual size of Validators greater than Determine Validator Size
	// then sort the Validators with VotingPower and return the most top Validators
	if validators.Size() > valSize {
		// Sort the Validator Set with Amount
		sort.Slice(validators.Validators, func(i, j int) bool {
			// Compare with remaining epoch first then, voting power
			if validators.Validators[i].RemainingEpoch == validators.Validators[j].RemainingEpoch {
				return validators.Validators[i].VotingPower.Cmp(validators.Validators[j].VotingPower) == 1
			} else {
				return validators.Validators[i].RemainingEpoch > validators.Validators[j].RemainingEpoch
			}
		})
		// Add knockout validator to refund list
		knockout := validators.Validators[valSize:]
		for _, k := range knockout {
			refund = append(refund, &tmTypes.RefundValidatorAmount{Address: common.BytesToAddress(k.Address), Amount: nil, Voteout: true})
		}

		validators.Validators = validators.Validators[:valSize]
	}

	return refund, nil
}

func (epoch *Epoch) GetEpochByBlockNumber(blockNumber uint64) *Epoch {

	if blockNumber >= epoch.StartBlock && blockNumber <= epoch.EndBlock {
		return epoch
	}

	for number := epoch.Number - 1; number >= 0; number-- {

		ep := loadOneEpoch(epoch.db, number, epoch.logger)
		if ep == nil {
			return nil
		}

		if blockNumber >= ep.StartBlock && blockNumber <= ep.EndBlock {
			return ep
		}
	}

	return nil
}

func (epoch *Epoch) Copy() *Epoch {
	return epoch.copy(true)
}

func (epoch *Epoch) copy(copyPrevNext bool) *Epoch {

	var previousEpoch, nextEpoch *Epoch
	if copyPrevNext {
		if epoch.previousEpoch != nil {
			previousEpoch = epoch.previousEpoch.copy(false)
		}

		if epoch.nextEpoch != nil {
			nextEpoch = epoch.nextEpoch.copy(false)
		}
	}

	return &Epoch{
		mtx:    epoch.mtx,
		db:     epoch.db,
		logger: epoch.logger,

		rs: epoch.rs,

		Number:           epoch.Number,
		RewardPerBlock:   new(big.Int).Set(epoch.RewardPerBlock),
		StartBlock:       epoch.StartBlock,
		EndBlock:         epoch.EndBlock,
		StartTime:        epoch.StartTime,
		EndTime:          epoch.EndTime,
		BlockGenerated:   epoch.BlockGenerated,
		Status:           epoch.Status,
		Validators:       epoch.Validators.Copy(),
		validatorVoteSet: epoch.validatorVoteSet.Copy(),

		previousEpoch: previousEpoch,
		nextEpoch:     nextEpoch,
	}
}

func (epoch *Epoch) estimateForNextEpoch(lastBlockHeight uint64, lastBlockTime time.Time) (rewardPerBlock *big.Int, blocksOfNextEpoch uint64) {

	var rewardFirstYear = epoch.rs.RewardFirstYear       //20000000e+18  every year
	var epochNumberPerYear = epoch.rs.EpochNumberPerYear //4380
	var totalYear = epoch.rs.TotalYear                   //10
	var timePerBlockOfEpoch int64

	const EMERGENCY_BLOCKS_OF_NEXT_EPOCH uint64 = 1000 // al least 1000 blocks per epoch
	const DefaultTimePerBlock int64 = 3000000000       // 3s

	zeroEpoch := loadOneEpoch(epoch.db, 0, epoch.logger)
	initStartTime := zeroEpoch.StartTime

	//from 0 year
	thisYear := epoch.Number / epochNumberPerYear
	nextYear := thisYear + 1

	log.Info("estimateForNextEpoch",
		"current epoch", epoch.Number,
		"epoch start block", epoch.StartBlock,
		"epoch end block", epoch.EndBlock,
		"epoch start time", epoch.StartTime,
		"epoch end time", epoch.EndTime,
		"last block height", lastBlockHeight,
		"rewardFirstYear", rewardFirstYear,
		"epochNumberPerYear", epochNumberPerYear,
		"totalYear", totalYear)

	// only use the current epoch to calculate the block time
	//if epoch.previousEpoch != nil {
	//	log.Infof("estimateForNextEpoch previous epoch, start time %v, end time %v", epoch.previousEpoch.StartTime.UnixNano(), epoch.previousEpoch.EndTime.UnixNano())
	//	prevEpoch := epoch.previousEpoch
	//	timePerBlockOfEpoch = prevEpoch.EndTime.Sub(prevEpoch.StartTime).Nanoseconds() / int64(prevEpoch.EndBlock-prevEpoch.StartBlock)
	//} else {
	timePerBlockOfEpoch = lastBlockTime.Sub(epoch.StartTime).Nanoseconds() / int64(lastBlockHeight-epoch.StartBlock)
	//}

	if timePerBlockOfEpoch <= 0 {
		log.Debugf("estimateForNextEpoch, timePerBlockOfEpoch is %v", timePerBlockOfEpoch)
		timePerBlockOfEpoch = DefaultTimePerBlock
	}

	epochLeftThisYear := epochNumberPerYear - epoch.Number%epochNumberPerYear - 1

	blocksOfNextEpoch = 0

	log.Info("estimateForNextEpoch",
		"epochLeftThisYear", epochLeftThisYear,
		"timePerBlockOfEpoch", timePerBlockOfEpoch)

	if epochLeftThisYear == 0 { //to another year

		nextYearStartTime := initStartTime.AddDate(int(nextYear), 0, 0)

		nextYearEndTime := nextYearStartTime.AddDate(1, 0, 0)

		timeLeftNextYear := nextYearEndTime.Sub(nextYearStartTime)

		epochLeftNextYear := epochNumberPerYear

		epochTimePerEpochLeftNextYear := timeLeftNextYear.Nanoseconds() / int64(epochLeftNextYear)

		blocksOfNextEpoch = uint64(epochTimePerEpochLeftNextYear / timePerBlockOfEpoch)

		log.Info("estimateForNextEpoch 0",
			"timePerBlockOfEpoch", timePerBlockOfEpoch,
			"nextYearStartTime", nextYearStartTime,
			"timeLeftNextYear", timeLeftNextYear,
			"epochLeftNextYear", epochLeftNextYear,
			"epochTimePerEpochLeftNextYear", epochTimePerEpochLeftNextYear,
			"blocksOfNextEpoch", blocksOfNextEpoch)

		if blocksOfNextEpoch < EMERGENCY_BLOCKS_OF_NEXT_EPOCH {
			blocksOfNextEpoch = EMERGENCY_BLOCKS_OF_NEXT_EPOCH //make it move ahead
			epoch.logger.Error("EstimateForNextEpoch Error: Please check the epoch_no_per_year setup in Genesis")
		}

		rewardPerEpochNextYear := calculateRewardPerEpochByYear(rewardFirstYear, int64(nextYear), int64(totalYear), int64(epochNumberPerYear))

		rewardPerBlock = new(big.Int).Div(rewardPerEpochNextYear, big.NewInt(int64(blocksOfNextEpoch)))

	} else {

		nextYearStartTime := initStartTime.AddDate(int(nextYear), 0, 0)

		timeLeftThisYear := nextYearStartTime.Sub(lastBlockTime)

		if timeLeftThisYear > 0 {

			epochTimePerEpochLeftThisYear := timeLeftThisYear.Nanoseconds() / int64(epochLeftThisYear)

			blocksOfNextEpoch = uint64(epochTimePerEpochLeftThisYear / timePerBlockOfEpoch)

			log.Info("estimateForNextEpoch 1",
				"timePerBlockOfEpoch", timePerBlockOfEpoch,
				"nextYearStartTime", nextYearStartTime,
				"timeLeftThisYear", timeLeftThisYear,
				"epochTimePerEpochLeftThisYear", epochTimePerEpochLeftThisYear,
				"blocksOfNextEpoch", blocksOfNextEpoch)
		}

		if blocksOfNextEpoch < EMERGENCY_BLOCKS_OF_NEXT_EPOCH {
			blocksOfNextEpoch = EMERGENCY_BLOCKS_OF_NEXT_EPOCH //make it move ahead
			epoch.logger.Error("EstimateForNextEpoch Error: Please check the epoch_no_per_year setup in Genesis")
		}

		log.Debugf("Current Epoch Number %v, This Year %v, Next Year %v, Epoch No Per Year %v, Epoch Left This year %v\n"+
			"initStartTime %v ; nextYearStartTime %v\n"+
			"Time Left This year %v, timePerBlockOfEpoch %v, blocksOfNextEpoch %v\n", epoch.Number, thisYear, nextYear, epochNumberPerYear, epochLeftThisYear, initStartTime, nextYearStartTime, timeLeftThisYear, timePerBlockOfEpoch, blocksOfNextEpoch)

		rewardPerEpochThisYear := calculateRewardPerEpochByYear(rewardFirstYear, int64(thisYear), int64(totalYear), int64(epochNumberPerYear))

		rewardPerBlock = new(big.Int).Div(rewardPerEpochThisYear, big.NewInt(int64(blocksOfNextEpoch)))

	}
	return rewardPerBlock, blocksOfNextEpoch
}

func calculateRewardPerEpochByYear(rewardFirstYear *big.Int, year, totalYear, epochNumberPerYear int64) *big.Int {
	if year > totalYear {
		return big.NewInt(0)
	}

	return new(big.Int).Div(rewardFirstYear, big.NewInt(epochNumberPerYear))
}

func (epoch *Epoch) Equals(other *Epoch, checkPrevNext bool) bool {

	if (epoch == nil && other != nil) || (epoch != nil && other == nil) {
		return false
	}

	if epoch == nil && other == nil {
		log.Debugf("Epoch equals epoch %v, other %v", epoch, other)
		return true
	}

	if !(epoch.Number == other.Number && epoch.RewardPerBlock.Cmp(other.RewardPerBlock) == 0 &&
		epoch.StartBlock == other.StartBlock && epoch.EndBlock == other.EndBlock &&
		epoch.Validators.Equals(other.Validators)) {
		return false
	}

	if checkPrevNext {
		if !epoch.previousEpoch.Equals(other.previousEpoch, false) ||
			!epoch.nextEpoch.Equals(other.nextEpoch, false) {
			return false
		}
	}
	log.Debugf("Epoch equals end, no matching")
	return true
}

func (epoch *Epoch) String() string {
	return fmt.Sprintf("Epoch : {"+
		"Number : %v,\n"+
		"RewardPerBlock : %v,\n"+
		"StartBlock : %v,\n"+
		"EndBlock : %v,\n"+
		"StartTime : %v,\n"+
		"EndTime : %v,\n"+
		"BlockGenerated : %v,\n"+
		"Status : %v,\n"+
		"Next Epoch : %v,\n"+
		"Prev Epoch : %v,\n"+
		"Contains RS : %v, \n"+
		"}",
		epoch.Number,
		epoch.RewardPerBlock,
		epoch.StartBlock,
		epoch.EndBlock,
		epoch.StartTime,
		epoch.EndTime,
		epoch.BlockGenerated,
		epoch.Status,
		epoch.nextEpoch,
		epoch.previousEpoch,
		epoch.rs != nil,
	)
}

func UpdateEpochEndTime(db dbm.DB, epNumber uint64, endTime time.Time) {
	// Load Epoch from DB
	ep := loadOneEpoch(db, epNumber, nil)
	if ep != nil {
		ep.mtx.Lock()
		defer ep.mtx.Unlock()
		// Set End Time
		ep.EndTime = endTime
		// Save back to DB
		db.SetSync(calcEpochKeyWithHeight(epNumber), ep.Bytes())
	}
}

//func (epoch *Epoch) GetForbiddenDuration() time.Duration {
//	return ForbiddenDuration
//}

// Update validator block time and set forbidden if this validator did not participate in consensus in one epoch
//func (epoch *Epoch) UpdateForbiddenState(header *types.Header, prevHeader *types.Header, commit *tmTypes.Commit, state *state.StateDB) {
//	validators := epoch.Validators.Validators
//	height := header.Number.Uint64()
//	//forbiddenTime := prevHeader.Time
//
//	//epoch.logger.Infof("Update validator forbidden state height %v", height)
//
//	if height <= 1 || height == epoch.StartBlock {
//		return
//	} else if height == epoch.EndBlock {
//		epoch.logger.Debugf("Update validator forbidden state, epoch end block %v", height)
//		// epoch end block set all validators mined block times 0
//		for _, v := range validators {
//			addr := common.BytesToAddress(v.Address[:])
//			state.SetMinedBlocks(addr, common.Big0)
//		}
//	} else if height == (epoch.EndBlock - 1) {
//		epoch.logger.Debugf("Update validator forbidden state, epoch end block - 1 %v", height)
//		// epoch end block - 1, if mined block times 0, set the validator forbidden true
//		for _, v := range validators {
//			addr := common.BytesToAddress(v.Address[:])
//			times := state.GetMinedBlocks(addr)
//			if times.Cmp(common.Big0) == 0 {
//				epoch.logger.Debugf("Update validator forbidden state, set %v forbidden, mined blocks %v, forbidden epoch %v", addr, times, ForbiddenEpoch)
//				state.SetForbidden(addr, true)
//				state.SetForbiddenTime(addr, ForbiddenEpoch)
//
//				state.MarkAddressForbidden(addr)
//			}
//		}
//	} else {
//		if commit == nil || commit.BitArray == nil {
//			epoch.logger.Debugf("Update validator forbidden state seenCommit %v", commit)
//			return
//		}
//
//		// update the mined block times
//		bitMap := commit.BitArray
//		for i := uint64(0); i < bitMap.Size(); i++ {
//			if bitMap.GetIndex(i) {
//				addr := common.BytesToAddress(validators[i].Address)
//				//vObj := state.GetOrNewStateObject(addr)
//
//				times := state.GetMinedBlocks(addr)
//				newTimes := big.NewInt(0)
//				newTimes.Add(times, common.Big1)
//
//				state.SetMinedBlocks(addr, newTimes)
//
//				//epoch.logger.Debugf("Update validator forbidden state, %v new mined block times %v, current times %v", addr, newTimes, times)
//			}
//		}
//	}
//
//}
