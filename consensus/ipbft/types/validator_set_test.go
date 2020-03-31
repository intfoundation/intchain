package types

import (
	"encoding/base64"
	"fmt"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/go-wire"
	"math/big"
	"testing"
)

var extraTreeVal = "0x0108696e74636861696e000000000003713d15d1c7c6c2488b8000000000000000000020011415d14546048c69022a12e0c8605338e68a841fb30114ec299e752fd4b41f0ea3d5a4e2b350e6f3a1471b010114ca95018425ce58d2e041eb5d9eda115f181214f800000000000000010114f4fd5d4e44794f67fa797f7824dd706934970e3c000000000003713d010201405678a97feace6079d60753d636ad7358349fcd2b222e1ffff393359f2c272e5d132f435b18967010e881db0dc52055c8c50ff0ac660195c494aa85a7015c7de30100000000000000030101000000000000000300"

type val struct {
	Address     string `json:"address"`
	PublicKey   string `json:"public_key"`
	RemainEpoch string `json:"remain_epoch"`
	VotingPower string `json:"voting_power"`
}

var valList = []*val{
	{
		Address:     "INT3CpFuk2cJ1te9WZV1w8Y3wkQCcA5Z",
		PublicKey:   "0x3589CA45DFD8EFB4103F8FB59BD3185B0BF73DC1EB8B659B0DDB2AFDEDB2BC433D89AF7791019456C79943113CA1370C9CA0B9DC341D884C5DDEDF0056E4057C45EB54DD9007917007C804BFBA3DF6E29D59CC43EFC7C3C10057C99818AEBECC21D8466741547D7CFCB709AF6293C338701EA82B2E4FC8F03188A47F35F261C6",
		RemainEpoch: "0x0",
		VotingPower: "0x2540be400",
	},
	{
		Address:     "INT38cqij9EtpxCZk9wrRY8u5U9reZAp",
		PublicKey:   "0x49DC3C606826D11FB38463197E3184DA388D829FC95835D4891C42CDCD0AACA84E65220F3034BF2B3C83497D90432C4A1B960F255613B062DFD3BD36CD9A061788EF399B142AAA8E3852206E72FF525DB76FBBDCC9F0FDB5C1C4028CE5D5421D3C334D893B67C90EC9DF4FD6D14C48F37DB133ACA561D4732D172DE3D41086F2",
		RemainEpoch: "0x0",
		VotingPower: "0x54b40b1f852bda000000",
	},
	{
		Address:     "INT3MUHiVzxaNdG1RAD7zQimzSZBtErX",
		PublicKey:   "0x372EC175F6D52B91FF43493AC9E113790C0CC6AD6562A1AD7E581717C40F017A6587F3FED9DFA590CED46AE37B99617F06DB3C36BE68F1AB9005276439AB44A455DDE3C1472467D84F851B9974D46A6A525E24638D5101BE207258D91983C03A2FB021C4C50DFE95E6169C698879ED9EA4A6089B24C596C17F46489823EA7CF7",
		RemainEpoch: "0x0",
		VotingPower: "0x54b40b1f852bda000000",
	},
}

func TestValidatorSet_Hash(t *testing.T) {
	var valSet ValidatorSet
	var validators []*Validator

	validators = make([]*Validator, len(valList))
	var totalVP = big.NewInt(0)
	for i, v := range valList {
		vp, _ := hexutil.DecodeBig(v.VotingPower)
		validators[i], _ = makeValidator(v.Address, v.PublicKey, vp)
		totalVP.Add(totalVP, vp)
	}
	// 用内部的生成 validatorset 的方法，会把 validator 重新排序，不应该使用
	validatorSet := NewValidatorSet(validators)

	valSet = ValidatorSet{
		validators,
		totalVP,
	}

	fmt.Printf("validatorset validators %v\n", valSet.Validators)
	fmt.Printf("validatorset validators %v\n", validatorSet.Validators)
	fmt.Printf("validatorset validatorset hash %v\n", valSet.Hash())
	fmt.Printf("validatorset validatorset hash hex %v\n\n", hexutil.Encode(valSet.Hash()))

	var extraData TendermintExtra
	bytes, err := hexutil.Decode(extraTreeVal)
	if err != nil {
		t.Fatal(err)
	}

	err = wire.ReadBinaryBytes(bytes, &extraData)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("extra data %v\n\n", extraData)

	talliedVotingPower, err := valSet.TalliedVotingPower(extraData.SeenCommit.BitArray)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("talliedVotingPower %v\n", talliedVotingPower)

	aggr, _ := valSet.GetAggrPubKeyAndAddress(extraData.SeenCommit.BitArray)
	for _, v := range aggr.PublicKeys {
		fmt.Printf("validator publickey %v\n", v)
	}

}

func makeValidator(address, blsPubKey string, vp *big.Int) (validator *Validator, err error) {
	var blsPK crypto.BLSPubKey
	blsPubKeyByte, err := hexutil.Decode(blsPubKey)
	if err != nil {
		return validator, err
	}

	for i, v := range blsPubKeyByte {
		blsPK[i] = v
	}

	validator = &Validator{
		Address:        []byte(address),
		PubKey:         blsPK,
		VotingPower:    vp,
		RemainingEpoch: uint64(0),
	}

	return validator, nil
}

func TestDecodeValidatorsHash(t *testing.T) {
	//var validatorsHashStr = "2uZsa8qZEBL29i1NtCx5xg48LkQ="
	var validatorsHashStr = "7CmedS/UtB8Oo9Wk4rNQ5vOhRxs="
	var validators Validator

	data, err := base64.StdEncoding.DecodeString(validatorsHashStr)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("validatorshash=%v\n", data)
	fmt.Printf("validatorshashhex=%v\n", hexutil.Encode(data))

	// 解不出来 validator
	err = wire.ReadBinaryBytes(data, &validators)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("validator=%v\n", validators)
}

func TestIntType(t *testing.T) {
	var v int
	v = 3
	s := v / 2
	fmt.Printf("s %v", s)
}
