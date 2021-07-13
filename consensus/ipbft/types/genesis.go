package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/common/hexutil"
	"math/big"
	"time"

	. "github.com/intfoundation/go-common"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/common"
)

//------------------------------------------------------------
// we store the gendoc in the db

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

var CONSENSUS_IPBFT string = "ipbft"

type GenesisValidator struct {
	EthAccount     common.Address `json:"address"`
	PubKey         crypto.PubKey  `json:"pub_key"`
	Amount         *big.Int       `json:"amount"`
	Name           string         `json:"name"`
	RemainingEpoch uint64         `json:"epoch"`
}

type GenesisCandidate struct {
	EthAccount common.Address `json:"address"`
	PubKey     crypto.PubKey  `json:"pub_key"`
}

type OneEpochDoc struct {
	Number         uint64             `json:"number"`
	RewardPerBlock *big.Int           `json:"reward_per_block"`
	StartBlock     uint64             `json:"start_block"`
	EndBlock       uint64             `json:"end_block"`
	Status         int                `json:"status"`
	Validators     []GenesisValidator `json:"validators"`
	//Candidates     []GenesisCandidate `json:"candidates"`
}

type RewardSchemeDoc struct {
	TotalReward        *big.Int `json:"total_reward"`
	RewardFirstYear    *big.Int `json:"reward_first_year"`
	EpochNumberPerYear uint64   `json:"epoch_no_per_year"`
	TotalYear          uint64   `json:"total_year"`
}

type GenesisDoc struct {
	ChainID      string          `json:"chain_id"`
	Consensus    string          `json:"consensus"`
	GenesisTime  time.Time       `json:"genesis_time"`
	RewardScheme RewardSchemeDoc `json:"reward_scheme"`
	CurrentEpoch OneEpochDoc     `json:"current_epoch"`
}

// Utility method for saving GenensisDoc as JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes, err := json.MarshalIndent(genDoc, "", "\t")
	if err != nil {
		fmt.Println(err)
	}

	return WriteFile(file, genDocBytes, 0644)
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (genDoc *GenesisDoc, err error) {
	err = json.Unmarshal(jsonBlob, &genDoc)
	return
}

var MainnetGenesisJSON string = `{
	"chain_id": "intchain",
	"consensus": "ipbft",
	"genesis_time": "2021-06-15T10:03:48.89357+08:00",
	"reward_scheme": {
		"total_reward": "0xa56fa5b99019a5c8000000",
		"reward_first_year": "0x108b2a2c28029094000000",
		"epoch_no_per_year": "0x111c",
		"total_year": "0xa"
	},
	"current_epoch": {
		"number": "0x0",
		"reward_per_block": "0x1a675944a9a10838",
		"start_block": "0x0",
		"end_block": "0x960",
		"validators": [
			{
				"address": "0x215720eba7f952fa579874a0b4925bbe8a39a82a",
				"pub_key": "0x6104995D7ED88ED04CCF69539DFC2FFD80F0AAAF723CDDC5A42B1853EF99868455B3F1BD6928C5958F813EE2FB916516FEF35D5CF185E9DEF85D96C0546C9F3981542080E17F6E0AA72930C48C1AC7F1A24D238802B27CBF46151B40BC1859AD6CCD9696C1FA83B4ADC31A14DF6DBEEFE236366638E2D3FA553CADD434E7F5B6",
				"amount": "0x54b40b1f852bda000000",
				"name": "",
				"epoch": "0x0"
			}
		]
	}
}`

var TestnetGenesisJSON string = `{
	"chain_id": "testnet",
	"consensus": "ipbft",
	"genesis_time": "2021-07-13T09:54:28.539144+08:00",
	"reward_scheme": {
		"total_reward": "0xa56fa5b99019a5c8000000",
		"reward_first_year": "0x108b2a2c28029094000000",
		"epoch_no_per_year": "0x111c",
		"total_year": "0xa"
	},
	"current_epoch": {
		"number": "0x0",
		"reward_per_block": "0x1a675944a9a10838",
		"start_block": "0x0",
		"end_block": "0x960",
		"validators": [
			{
				"address": "0x9fe77cb625d2aa7a974b4ee88c9095893da8c1a0",
				"pub_key": "0x15A3CA6AC1DB28C49770C619AA1CD5D2C80F2B5A0397A26469874D060D31648048023C104ACD2F26B30BA7A9378B99A075EC6696BB8DF5EB36E878A18F4A7DF715E33BD07F8EA0175AA308DD438E164D904AA081C77555F9DF280C485BC0C8296A339B418D06D8CB91905F2C07031F72B00B59CB3B2A8CF57412645C41C4557A",
				"amount": "0x54b40b1f852bda000000",
				"name": "",
				"epoch": "0x0"
			}
		]
	}
}
`

func (ep OneEpochDoc) MarshalJSON() ([]byte, error) {
	type hexEpoch struct {
		Number         hexutil.Uint64     `json:"number"`
		RewardPerBlock *hexutil.Big       `json:"reward_per_block"`
		StartBlock     hexutil.Uint64     `json:"start_block"`
		EndBlock       hexutil.Uint64     `json:"end_block"`
		Validators     []GenesisValidator `json:"validators"`
	}
	var enc hexEpoch
	enc.Number = hexutil.Uint64(ep.Number)
	enc.RewardPerBlock = (*hexutil.Big)(ep.RewardPerBlock)
	enc.StartBlock = hexutil.Uint64(ep.StartBlock)
	enc.EndBlock = hexutil.Uint64(ep.EndBlock)
	if ep.Validators != nil {
		enc.Validators = ep.Validators
	}
	return json.Marshal(&enc)
}

func (ep *OneEpochDoc) UnmarshalJSON(input []byte) error {
	type hexEpoch struct {
		Number         hexutil.Uint64     `json:"number"`
		RewardPerBlock *hexutil.Big       `json:"reward_per_block"`
		StartBlock     hexutil.Uint64     `json:"start_block"`
		EndBlock       hexutil.Uint64     `json:"end_block"`
		Validators     []GenesisValidator `json:"validators"`
	}
	var dec hexEpoch
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	ep.Number = uint64(dec.Number)
	ep.RewardPerBlock = (*big.Int)(dec.RewardPerBlock)
	ep.StartBlock = uint64(dec.StartBlock)
	ep.EndBlock = uint64(dec.EndBlock)
	if dec.Validators == nil {
		return errors.New("missing required field 'validators' for Genesis/epoch")
	}
	ep.Validators = dec.Validators
	return nil
}

func (gv GenesisValidator) MarshalJSON() ([]byte, error) {
	type hexValidator struct {
		Address        common.Address `json:"address"`
		PubKey         string         `json:"pub_key"`
		Amount         *hexutil.Big   `json:"amount"`
		Name           string         `json:"name"`
		RemainingEpoch hexutil.Uint64 `json:"epoch"`
	}
	var enc hexValidator
	enc.Address = gv.EthAccount
	enc.PubKey = gv.PubKey.KeyString()
	enc.Amount = (*hexutil.Big)(gv.Amount)
	enc.Name = gv.Name
	enc.RemainingEpoch = hexutil.Uint64(gv.RemainingEpoch)

	return json.Marshal(&enc)
}

func (gv *GenesisValidator) UnmarshalJSON(input []byte) error {
	type hexValidator struct {
		Address        common.Address `json:"address"`
		PubKey         string         `json:"pub_key"`
		Amount         *hexutil.Big   `json:"amount"`
		Name           string         `json:"name"`
		RemainingEpoch hexutil.Uint64 `json:"epoch"`
	}
	var dec hexValidator
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	gv.EthAccount = dec.Address

	pubkeyBytes := common.FromHex(dec.PubKey)
	if dec.PubKey == "" || len(pubkeyBytes) != 128 {
		return errors.New("wrong format of required field 'pub_key' for Genesis/epoch/validators")
	}
	var blsPK crypto.BLSPubKey
	copy(blsPK[:], pubkeyBytes)
	gv.PubKey = blsPK

	if dec.Amount == nil {
		return errors.New("missing required field 'amount' for Genesis/epoch/validators")
	}
	gv.Amount = (*big.Int)(dec.Amount)
	gv.Name = dec.Name
	gv.RemainingEpoch = uint64(dec.RemainingEpoch)
	return nil
}

func (rs RewardSchemeDoc) MarshalJSON() ([]byte, error) {
	type hexRewardScheme struct {
		TotalReward        *hexutil.Big   `json:"total_reward"`
		RewardFirstYear    *hexutil.Big   `json:"reward_first_year"`
		EpochNumberPerYear hexutil.Uint64 `json:"epoch_no_per_year"`
		TotalYear          hexutil.Uint64 `json:"total_year"`
	}
	var enc hexRewardScheme
	enc.TotalReward = (*hexutil.Big)(rs.TotalReward)
	enc.RewardFirstYear = (*hexutil.Big)(rs.RewardFirstYear)
	enc.EpochNumberPerYear = hexutil.Uint64(rs.EpochNumberPerYear)
	enc.TotalYear = hexutil.Uint64(rs.TotalYear)

	return json.Marshal(&enc)
}

func (rs *RewardSchemeDoc) UnmarshalJSON(input []byte) error {
	type hexRewardScheme struct {
		TotalReward        *hexutil.Big   `json:"total_reward"`
		RewardFirstYear    *hexutil.Big   `json:"reward_first_year"`
		EpochNumberPerYear hexutil.Uint64 `json:"epoch_no_per_year"`
		TotalYear          hexutil.Uint64 `json:"total_year"`
	}
	var dec hexRewardScheme
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.TotalReward == nil {
		return errors.New("missing required field 'total_reward' for Genesis/reward_scheme")
	}
	rs.TotalReward = (*big.Int)(dec.TotalReward)
	if dec.RewardFirstYear == nil {
		return errors.New("missing required field 'reward_first_year' for Genesis/reward_scheme")
	}
	rs.RewardFirstYear = (*big.Int)(dec.RewardFirstYear)

	rs.EpochNumberPerYear = uint64(dec.EpochNumberPerYear)
	rs.TotalYear = uint64(dec.TotalYear)

	return nil
}
