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
				"address": "0x9038d5c22f7be9a1447e92b3be2599facb690988",
				"pub_key": "0x196A0CCFB2C205066F6C2373BB84A0C81AF4A22A8B6541172D79A236C572322871F851EE52B56B8FB0D7B9EC61B18E236C7E1A5D726AD4DC9617D4EFCADB152B481F2085038875F6C7505B7AD2DE0C2F03B219D0CC4688A8A9B28CA0A4E45C624195458FD750A0D02D9F6E544BB006883429DC61BA39035C10CD10467C828A1A",
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
	"genesis_time": "2021-06-17T16:53:06.642736+08:00",
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
				"address": "0x3e08a3e62001ef7c45629bddb0cd6b10b4564fe3",
				"pub_key": "0x1B7BAE2C17B447AE86F5A13393A5C9C0A6CEF1BC585993B7A740DE95E462697162E53C81DEBFE70D2EB31CCA72C362FF8E4BEDA275CB47A0357E0529E6E202614D3B5761262D7E2AE1CA9EC2AC6F797A549E49C0F6AFC1E4E18C6FEACD0935923FBB01D192C66D207EB8CD13931C7E5E2D40131C2A433178E54EE9013943B5E5",
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
