package types

import (
	. "github.com/intfoundation/go-common"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"time"
)

type EpochApi struct {
	Number         hexutil.Uint64    `json:"number"`
	RewardPerBlock *hexutil.Big      `json:"rewardPerBlock"`
	StartBlock     hexutil.Uint64    `json:"startBlock"`
	EndBlock       hexutil.Uint64    `json:"endBlock"`
	StartTime      time.Time         `json:"startTime"`
	EndTime        time.Time         `json:"endTime"`
	Validators     []*EpochValidator `json:"validators"`
}

type EpochVotesApi struct {
	EpochNumber hexutil.Uint64           `json:"voteForEpoch"`
	StartBlock  hexutil.Uint64           `json:"startBlock"`
	EndBlock    hexutil.Uint64           `json:"endBlock"`
	Votes       []*EpochValidatorVoteApi `json:"votes"`
}

type EpochValidatorVoteApi struct {
	EpochValidator
	Salt     string      `json:"salt"`
	VoteHash common.Hash `json:"voteHash"` // VoteHash = Keccak256(Epoch Number + PubKey + Amount + Salt)
	TxHash   common.Hash `json:"txHash"`
}

type EpochValidator struct {
	Address        common.Address `json:"address"`
	PubKey         string         `json:"publicKey"`
	Amount         *hexutil.Big   `json:"votingPower"`
	RemainingEpoch hexutil.Uint64 `json:"remainEpoch"`
}

type EpochCandidate struct {
	Address common.Address `json:"address"`
}

type TendermintExtraApi struct {
	ChainID         string         `json:"chainId"`
	Height          hexutil.Uint64 `json:"height"`
	Time            time.Time      `json:"time"`
	NeedToSave      bool           `json:"needToSave"`
	NeedToBroadcast bool           `json:"needToBroadcast"`
	EpochNumber     hexutil.Uint64 `json:"epochNumber"`
	SeenCommitHash  string         `json:"lastCommitHash"` // commit from validators from the last block
	ValidatorsHash  string         `json:"validatorsHash"` // validators for the current block
	SeenCommit      *CommitApi     `json:"seenCommit"`
	EpochBytes      []byte         `json:"epochBytes"`
}

type CommitApi struct {
	BlockID BlockIDApi     `json:"blockID"`
	Height  hexutil.Uint64 `json:"height"`
	Round   int            `json:"round"`

	// BLS signature aggregation to be added here
	SignAggr crypto.BLSSignature `json:"signAggr"`
	BitArray *BitArray           `json:"bitArray"`

	//// Volatile
	//hash []byte
}

type BlockIDApi struct {
	Hash        string           `json:"hash"`
	PartsHeader PartSetHeaderApi `json:"parts"`
}

type PartSetHeaderApi struct {
	Total hexutil.Uint64 `json:"total"`
	Hash  string         `json:"hash"`
}

type ConsensusAggr struct {
	PublicKeys []string         `json:"publicKey"`
	Addresses  []common.Address `json:"address"`
}

//type ValidatorStatus struct {
//	IsForbidden bool `json:"isForbidden"`
//}

//type CandidateApi struct {
//	CandidateList []string `json:"candidateList"`
//}

//type ForbiddenApi struct {
//	ForbiddenList []string `json:"forbiddenList"`
//}
