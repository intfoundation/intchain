package types

import (
	"encoding/hex"
	"fmt"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/go-wire"
	"math/big"
	"testing"
	"time"
)

var extraHex = "0x0108696e74636861696e0000000000002c0615cd7d205898b7400000000000000000000201149d2b5214e24104f648819a2b851970e02e0f73c50114dae66c6bca991012f6f62d4db42c79c60e3c2e44010114980aaa734720bdc9d93e3d4030a135aec4910bac0000000000000001011415a43260d301c016d636ec117eac56578f1157990000000000002c06000140439243a34a9ef90323f7db46d7800b7049fa4a50599956054d716de1cf37f7dc686c8a5c6510a17609cda05dd53737cc8635f15da4f7f48ccc1c288db0960e730100000000000000010101000000000000000100"
var extraHex2 = "0x0108696e74636861696e000000000003713d15d1c7c6c2488b8000000000000000000020011415d14546048c69022a12e0c8605338e68a841fb30114ec299e752fd4b41f0ea3d5a4e2b350e6f3a1471b010114ca95018425ce58d2e041eb5d9eda115f181214f800000000000000010114f4fd5d4e44794f67fa797f7824dd706934970e3c000000000003713d010201405678a97feace6079d60753d636ad7358349fcd2b222e1ffff393359f2c272e5d132f435b18967010e881db0dc52055c8c50ff0ac660195c494aa85a7015c7de30100000000000000030101000000000000000300"
var blsPubkeyHex = "0x31A9A1B8808146B846E8D919F9BF565F3F18E1F5359101B05BB41A28761F671411D824C15794553ACC74E85E5B6E67919B7C4CCFB844AC21611DD9AD2B78AA2069B988F89E9C2C6DA57A68960584D0346FFD910823DE06C55D7B151AECBBC4731F76DDA5E7202BB1EAF824F065EB43E187D14CC1EBC5C69033D42D03F699D8F2"

func TestEncodeTendermintExtra(t *testing.T) {
	var extra = TendermintExtra{}
	extra = TendermintExtra{
		ChainID:         "intchain",
		Height:          uint64(11270),
		Time:            time.Now(),
		NeedToSave:      false,
		NeedToBroadcast: false,
		EpochNumber:     uint64(2),
		SeenCommitHash:  []byte{0x3e, 0x78, 0x73, 0xef, 0xaf, 0x71, 0x6, 0xda, 0x71, 0x62, 0x68, 0xbe, 0x31, 0xd2, 0x73, 0xf, 0xa4, 0x28, 0x35, 0xda},
		ValidatorsHash:  []byte{0xda, 0xe6, 0x6c, 0x6b, 0xca, 0x99, 0x10, 0x12, 0xf6, 0xf6, 0x2d, 0x4d, 0xb4, 0x2c, 0x79, 0xc6, 0xe, 0x3c, 0x2e, 0x44},
		SeenCommit:      &Commit{},
		EpochBytes:      []byte{},
	}
	extraHash := extra.Hash()
	fmt.Printf("extraHash=%v\n", hex.EncodeToString(extraHash))

}

func TestValidator_Hash(t *testing.T) {
	var blsPubKey crypto.BLSPubKey
	blsPubKeyByte, err := hexutil.Decode(blsPubkeyHex)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("blsPubKeyByte len=%v\n", len(blsPubKeyByte))
	for i, v := range blsPubKeyByte {
		blsPubKey[i] = v
	}
	fmt.Printf("blsPubKey%v\n", blsPubKey)
	validator := Validator{
		Address:        []byte("32H8py5Jg396p7QNDUwTwkeVod15ksxne5"),
		PubKey:         blsPubKey,
		VotingPower:    big.NewInt(10000000000),
		RemainingEpoch: uint64(0),
	}

	validatorHash := validator.Hash()
	fmt.Printf("validatorHash=%v\n", validatorHash)
	fmt.Printf("validatorHashHex=%v\n", hexutil.Encode(validatorHash))
}

//type val struct {
//	Address 	string `json:"address"`
//	PublicKey 	string `json:"public_key"`
//	RemainEpoch string `json:"remain_epoch"`
//	VotingPower string `json:"voting_power"`
//}
//func TestValidatorSet_Hash(t *testing.T) {
//	var valSet ValidatorSet
//	var validators []*Validator
//	var valList = []*val{
//		{
//			Address: "INT3CpFuk2cJ1te9WZV1w8Y3wkQCcA5Z",
//			PublicKey: "0x3589CA45DFD8EFB4103F8FB59BD3185B0BF73DC1EB8B659B0DDB2AFDEDB2BC433D89AF7791019456C79943113CA1370C9CA0B9DC341D884C5DDEDF0056E4057C45EB54DD9007917007C804BFBA3DF6E29D59CC43EFC7C3C10057C99818AEBECC21D8466741547D7CFCB709AF6293C338701EA82B2E4FC8F03188A47F35F261C6",
//			RemainEpoch:"0x0",
//			VotingPower: "0x2540be400",
//		},
//		{
//			Address: "INT38cqij9EtpxCZk9wrRY8u5U9reZAp",
//			PublicKey: "0x49DC3C606826D11FB38463197E3184DA388D829FC95835D4891C42CDCD0AACA84E65220F3034BF2B3C83497D90432C4A1B960F255613B062DFD3BD36CD9A061788EF399B142AAA8E3852206E72FF525DB76FBBDCC9F0FDB5C1C4028CE5D5421D3C334D893B67C90EC9DF4FD6D14C48F37DB133ACA561D4732D172DE3D41086F2",
//			RemainEpoch:"0x0",
//			VotingPower: "0x54b40b1f852bda000000",
//		},
//		//{
//		//	Address: "INT3MUHiVzxaNdG1RAD7zQimzSZBtErX",
//		//	PublicKey: "0x372EC175F6D52B91FF43493AC9E113790C0CC6AD6562A1AD7E581717C40F017A6587F3FED9DFA590CED46AE37B99617F06DB3C36BE68F1AB9005276439AB44A455DDE3C1472467D84F851B9974D46A6A525E24638D5101BE207258D91983C03A2FB021C4C50DFE95E6169C698879ED9EA4A6089B24C596C17F46489823EA7CF7",
//		//	RemainEpoch:"0x0",
//		//	VotingPower: "0x54b40b1f852bda000000",
//		//},
//	}
//
//	var totalVP = big.NewInt(0)
//	validators = make([]*Validator, len(valList))
//	for i, v := range valList {
//		vp, _ := hexutil.DecodeBig(v.VotingPower)
//		validators[i], _ = makeValidator(v.Address, v.PublicKey, vp)
//		totalVP.Add(totalVP, vp)
//	}
//
//	valSet = ValidatorSet{
//		validators,
//		totalVP,
//	}
//
//	fmt.Printf("validatorset validators %v\n", valSet.Validators)
//	fmt.Printf("validatorset validatorset hash %v\n", valSet.Hash())
//	fmt.Printf("validatorset validatorset hash hex %v\n", hexutil.Encode(valSet.Hash()))
//
//}

func TestDecodeTendermintExtra(t *testing.T) {
	var extra = TendermintExtra{}
	extraByte, err := hexutil.Decode(extraHex2)
	fmt.Printf("extraByte=%v\n\n", extraByte)
	if err != nil {
		t.Error(err)
	}
	err = wire.ReadBinaryBytes(extraByte, &extra)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("extra=%v\n", extra)
	fmt.Printf("extra.validatorshash=%v\n", extra.ValidatorsHash)
	fmt.Printf("extra.validatorshashhex=%v\n\n", hexutil.Encode(extra.ValidatorsHash))
	fmt.Printf("extra.SeenCommit=%v\n", *extra.SeenCommit)
	fmt.Printf("extra.SeenCommit.BitArray=%v\n", extra.SeenCommit.BitArray)
	fmt.Printf("extra.SeenCommit.NumCommits=%v\n", extra.SeenCommit.NumCommits())
}

//func TestDecodeValidatorsHash(t *testing.T) {
//	//var validatorsHashStr = "2uZsa8qZEBL29i1NtCx5xg48LkQ="
//	var validatorsHashStr = "7CmedS/UtB8Oo9Wk4rNQ5vOhRxs="
//	var validators Validator
//
//	data, err := base64.StdEncoding.DecodeString(validatorsHashStr)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("validatorshash=%v\n", data)
//	fmt.Printf("validatorshashhex=%v\n", hexutil.Encode(data))
//
//	// 解不出来 validator
//	err = wire.ReadBinaryBytes(data, &validators)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("validator=%v\n", validators)
//}

//func makeValidator(address, blsPubKey string, vp *big.Int) (validator *Validator, err error){
//	var blsPK crypto.BLSPubKey
//	blsPubKeyByte, err := hexutil.Decode(blsPubKey)
//	if err != nil {
//		return validator, err
//	}
//
//	for i, v := range blsPubKeyByte {
//		blsPK[i] = v
//	}
//
//	validator = &Validator{
//		Address:        []byte(address),
//		PubKey:         blsPK,
//		VotingPower:    vp,
//		RemainingEpoch: uint64(0),
//	}
//
//	return validator, nil
//}
