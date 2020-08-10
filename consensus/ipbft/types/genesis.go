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

var CONSENSUS_POS string = "pos"
var CONSENSUS_POW string = "pow"
var CONSENSUS_IPBFT string = "ipbft"

type GenesisValidator struct {
	EthAccount     common.Address `json:"address"`
	PubKey         crypto.PubKey  `json:"pub_key"`
	Amount         *big.Int       `json:"amount"`
	Name           string         `json:"name"`
	RemainingEpoch uint64         `json:"epoch"`
}

type OneEpochDoc struct {
	Number         uint64             `json:"number"`
	RewardPerBlock *big.Int           `json:"reward_per_block"`
	StartBlock     uint64             `json:"start_block"`
	EndBlock       uint64             `json:"end_block"`
	Status         int                `json:"status"`
	Validators     []GenesisValidator `json:"validators"`
}

type RewardSchemeDoc struct {
	TotalReward        *big.Int `json:"total_reward"`
	RewardFirstYear    *big.Int `json:"reward_first_year"`
	EpochNumberPerYear uint64   `json:"epoch_no_per_year"`
	TotalYear          uint64   `json:"total_year"`
}

type GenesisDoc struct {
	ChainID      string          `json:"chain_id"`
	Consensus    string          `json:"consensus"` //should be 'pos' or 'pow'
	GenesisTime  time.Time       `json:"genesis_time"`
	RewardScheme RewardSchemeDoc `json:"reward_scheme"`
	CurrentEpoch OneEpochDoc     `json:"current_epoch"`
}

// 写入文件使用
type GenesisDocWrite struct {
	ChainID      string           `json:"chain_id"`
	Consensus    string           `json:"consensus"` //should be 'pos' or 'pow'
	GenesisTime  time.Time        `json:"genesis_time"`
	RewardScheme RewardSchemeDoc  `json:"reward_scheme"`
	CurrentEpoch OneEpochDocWrite `json:"current_epoch"`
}

// 写入文件使用
type OneEpochDocWrite struct {
	Number         uint64                  `json:"number"`
	RewardPerBlock *big.Int                `json:"reward_per_block"`
	StartBlock     uint64                  `json:"start_block"`
	EndBlock       uint64                  `json:"end_block"`
	Status         int                     `json:"status"`
	Validators     []GenesisValidatorWrite `json:"validators"`
}

// 写入文件使用
type GenesisValidatorWrite struct {
	EthAccount     string        `json:"address"`
	PubKey         crypto.PubKey `json:"pub_key"`
	Amount         *big.Int      `json:"amount"`
	Name           string        `json:"name"`
	RemainingEpoch uint64        `json:"epoch"`
}

// Utility method for saving GenensisDoc as JSON file.
//func (genDoc *GenesisDoc) SaveAs(file string) error {
//	genDocBytes, err := json.MarshalIndent(genDoc, "", "\t")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	return WriteFile(file, genDocBytes, 0644)
//}
// 写入文件使用
func (genDoc *GenesisDoc) SaveAs(file string) error {

	genDocWrite := GenesisDocWrite{
		ChainID:      genDoc.ChainID,
		Consensus:    genDoc.Consensus,
		GenesisTime:  genDoc.GenesisTime,
		RewardScheme: genDoc.RewardScheme,
		CurrentEpoch: OneEpochDocWrite{
			Number:         genDoc.CurrentEpoch.Number,
			RewardPerBlock: genDoc.CurrentEpoch.RewardPerBlock,
			StartBlock:     genDoc.CurrentEpoch.StartBlock,
			EndBlock:       genDoc.CurrentEpoch.EndBlock,
			Status:         genDoc.CurrentEpoch.Status,
			Validators:     make([]GenesisValidatorWrite, len(genDoc.CurrentEpoch.Validators)),
		},
	}
	for i, v := range genDoc.CurrentEpoch.Validators {
		genDocWrite.CurrentEpoch.Validators[i] = GenesisValidatorWrite{
			EthAccount:     v.EthAccount.String(),
			PubKey:         v.PubKey,
			Amount:         v.Amount,
			Name:           v.Name,
			RemainingEpoch: v.RemainingEpoch,
		}
	}

	genDocBytes, err := json.MarshalIndent(genDocWrite, "", "\t")
	if err != nil {
		fmt.Println(err)
	}

	return WriteFile(file, genDocBytes, 0644)
}

//------------------------------------------------------------
// Make genesis state from file

//func GenesisDocFromJSON(jsonBlob []byte) (genDoc *GenesisDoc, err error) {
//	err = json.Unmarshal(jsonBlob, &genDoc)
//	return
//}

// 读取 genesisdocjson，并做转换
func GenesisDocFromJSON(jsonBlob []byte) (genDoc *GenesisDoc, err error) {
	var genDocWrite *GenesisDocWrite
	err = json.Unmarshal(jsonBlob, &genDocWrite)
	if err != nil {
		return &GenesisDoc{}, err
	}

	genDoc = &GenesisDoc{
		ChainID:      genDocWrite.ChainID,
		Consensus:    genDocWrite.Consensus,
		GenesisTime:  genDocWrite.GenesisTime,
		RewardScheme: genDocWrite.RewardScheme,
		CurrentEpoch: OneEpochDoc{
			Number:         genDocWrite.CurrentEpoch.Number,
			RewardPerBlock: genDocWrite.CurrentEpoch.RewardPerBlock,
			StartBlock:     genDocWrite.CurrentEpoch.StartBlock,
			EndBlock:       genDocWrite.CurrentEpoch.EndBlock,
			Status:         genDocWrite.CurrentEpoch.Status,
			Validators:     make([]GenesisValidator, len(genDocWrite.CurrentEpoch.Validators)),
		},
	}
	for i, v := range genDocWrite.CurrentEpoch.Validators {
		genDoc.CurrentEpoch.Validators[i] = GenesisValidator{
			EthAccount:     common.StringToAddress(v.EthAccount),
			PubKey:         v.PubKey,
			Amount:         v.Amount,
			Name:           v.Name,
			RemainingEpoch: v.RemainingEpoch,
		}
	}

	return
}

var MainnetGenesisJSON string = `{
	"chain_id": "intchain",
	"consensus": "ipbft",
	"genesis_time": "2020-08-06T16:06:17.107216+08:00",
	"reward_scheme": {
		"total_reward": "0xa56fa5b99019a5c8000000",
		"reward_first_year": "0x108b2a2c28029094000000",
		"epoch_no_per_year": "0x111c",
		"total_year": "0xa"
	},
	"current_epoch": {
		"number": "0x0",
		"reward_per_block": "0x8cd1dc18de05834",
		"start_block": "0x0",
		"end_block": "0x1c20",
		"validators": [
			{
				"address": "INT33snTw1SYrAyzSmuKj9NFxTx5iF2L",
				"pub_key": "0x691BA6C071CA474C7EB7953823D78C8F780C58F626F87583A05DDA1DE9FABE6B122F68FF7BB7845FFBEDA1021833E49F5E2363AA06DC14FD70D1ADE804B8103C31C0D11A0691EF0C01974FE01699DA22E0C7A560A102AF4FA2B299BD67D6CD2941E8E0631A14D758C535BB4CA7CCF08A314536723FF6859F0A67FE58E53C4E81",
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
	"genesis_time": "2020-08-10T14:37:08.45674+08:00",
	"reward_scheme": {
		"total_reward": "0xa56fa5b99019a5c8000000",
		"reward_first_year": "0x108b2a2c28029094000000",
		"epoch_no_per_year": "0x111c",
		"total_year": "0xa"
	},
	"current_epoch": {
		"number": "0x0",
		"reward_per_block": "0x8cd1dc18de05834",
		"start_block": "0x0",
		"end_block": "0x1c20",
		"validators": [
			{
				"address": "INT3CFVNpTwr3QrykhPWiLP8n9wsyCVa",
				"pub_key": "0x1920843C3373C7C0D0C9C772CDD4E581FBC6A34B1EFC586FF78C4169CAEC22096B91C1CA4D8E5DB90437C23EAAC93031D408B14E6B777F5745D2B56D6B2F503A413E97C4445889C338D4F31982ECF43326140ECF2404DB40DB914949720C74078BFC302A3CC693A2ED57A9EE776300C56367A7173DD26B83F956C90690AD5973",
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

// 写入文件中间转换
func (ep OneEpochDocWrite) MarshalJSON() ([]byte, error) {
	type hexEpoch struct {
		Number         hexutil.Uint64          `json:"number"`
		RewardPerBlock *hexutil.Big            `json:"reward_per_block"`
		StartBlock     hexutil.Uint64          `json:"start_block"`
		EndBlock       hexutil.Uint64          `json:"end_block"`
		Validators     []GenesisValidatorWrite `json:"validators"`
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

// 写入文件中间转换
func (ep *OneEpochDocWrite) UnmarshalJSON(input []byte) error {
	type hexEpoch struct {
		Number         hexutil.Uint64          `json:"number"`
		RewardPerBlock *hexutil.Big            `json:"reward_per_block"`
		StartBlock     hexutil.Uint64          `json:"start_block"`
		EndBlock       hexutil.Uint64          `json:"end_block"`
		Validators     []GenesisValidatorWrite `json:"validators"`
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

// 写入文件中间转换
func (gv GenesisValidatorWrite) MarshalJSON() ([]byte, error) {
	type hexValidator struct {
		Address        string         `json:"address"`
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

// 写入文件中间转换
func (gv *GenesisValidatorWrite) UnmarshalJSON(input []byte) error {
	type hexValidator struct {
		Address        string         `json:"address"`
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
