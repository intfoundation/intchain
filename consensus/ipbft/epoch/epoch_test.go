package epoch

import (
	"fmt"
	"github.com/intfoundation/intchain/common"
	tmTypes "github.com/intfoundation/intchain/consensus/ipbft/types"
	"math/big"
	"sort"
	"testing"
	"time"
)

func TestEstimateEpoch(t *testing.T) {
	timeA := time.Now()
	timeB := timeA.Unix()
	timeStr := timeA.String()
	t.Logf("\ntimeA now: %v\n, timeA unix: %v\n, timeA String: %v\n, timeA second %v\n", timeA, timeB, timeStr, timeA.Second())

	formatTimeStr := "2020-06-23 09:51:38.397502 +0800 CST m=+10323.270024761"
	parse, e := time.Parse("", formatTimeStr)
	if e == nil {
		t.Logf("time: %v", parse)
	} else {
		t.Errorf("parse error: %v", e)
	}

	timeC := time.Now().UnixNano()
	t.Logf("time c %v", timeC)
	t.Logf("time c %v", timeC)

	var timeD time.Time
	t.Logf("t %v", timeD) // 0001-01-01 00:00:00 +0000 UTC

	fee := big.NewInt(30000000)
	halfFee := big.NewInt(0).Div(fee, big.NewInt(2))
	t.Logf("half fee %v, fee %v\n", halfFee, fee)
}

func TestVoteSetCompare(t *testing.T) {
	var voteArr []*EpochValidatorVote
	voteArr = []*EpochValidatorVote{
		{
			Address: common.StringToAddress("INT3CFVNpTwr3QrykhPWiLP8n9wsyCVa"),
			Amount:  big.NewInt(1),
		},
		{
			Address: common.StringToAddress("INT39iewq2jAyREvwqAZX4Wig5GVmSsc"),
			Amount:  big.NewInt(1),
		},
		{
			Address: common.StringToAddress("INT3JqvEfW7eTymfA6mfruwipcc1dAEi"),
			Amount:  big.NewInt(1),
		},
		{
			Address: common.StringToAddress("INT3D4sNnoM4NcLJeosDKUjxgwhofDdi"),
			Amount:  big.NewInt(1),
		},
		{
			Address: common.StringToAddress("INT3ETpxfNquuFa2czSHuFJTyhuepgXa"),
			Amount:  big.NewInt(1),
		},
		{
			Address: common.StringToAddress("INT3MjFkyK3bZ6oSCK8i38HVxbbsiRTY"),
			Amount:  big.NewInt(1),
		},
	}

	sort.Slice(voteArr, func(i, j int) bool {
		if voteArr[i].Amount.Cmp(voteArr[j].Amount) == 0 {
			return compareAddress(voteArr[i].Address[:], voteArr[j].Address[:])
		}

		return voteArr[i].Amount.Cmp(voteArr[j].Amount) == 1
	})
	for i := range voteArr {
		fmt.Printf("address:%v, amount: %v\n", voteArr[i].Address, voteArr[i].Amount)
	}
}

func TestVoteSetRemove(t *testing.T) {
	var validatorsArr []*tmTypes.Validator
	validatorsArr = []*tmTypes.Validator{
		{
			Address:     common.HexToAddress("0x2b14a6b2649a28b5fc90c42bf90f5242ea82f66a").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x6784d4990a5b042f17f149a387bac8e2f6f74064").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xeae4528f182e96ce021a8b803ce94755d65c1779").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xc906ae8ac16b80c2b591bde248283b66974756ea").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xc6540804d9994f6642d4786468c2eef0c66f69aa").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x11668e7a9ef9aa00fffaa4394c848226a588c860").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x17d28a143d48f325e5375fe54e427178e0ae5945").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xd5b9a960badecc32b67c89e04a7174a6880f7199").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x9ddc9c979f8f8d3d98601b05c4f71b504663640a").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x4fb47b878d1ebaa9e838295d305c928783c3442f").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x881f9437d5488a0162e2afacbc124e45fb24a527").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xe51116ef1cc8917bb9ef1b36c73c0fa79062746e").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0x56db076b7d71d2b3a4fcabbe9df2d3a06e5611ec").Bytes(),
			VotingPower: big.NewInt(1),
		},
		{
			Address:     common.HexToAddress("0xe9e13d382f366692bedda1d18c000a18180d410c").Bytes(),
			VotingPower: big.NewInt(1),
		},
	}

	validatorSet := tmTypes.NewValidatorSet(validatorsArr)
	fmt.Printf("validator set: %v\n", validatorSet)
	for _, val := range validatorSet.Validators {
		addr := common.BytesToAddress(val.Address)
		t.Logf("address: %X", addr)
		t.Logf("len 1: %v, %v", len(validatorSet.Validators), cap(validatorSet.Validators))
		t.Logf("ten address: %X", common.BytesToAddress(validatorSet.Validators[10].Address))
		if addr == common.HexToAddress("0x17d28a143d48f325e5375fe54e427178e0ae5945") ||
			addr == common.HexToAddress("0x56db076b7d71d2b3a4fcabbe9df2d3a06e5611ec") ||
			addr == common.HexToAddress("0xe9e13d382f366692bedda1d18c000a18180d410c") {
			validatorSet.Remove(val.Address)
			//fmt.Printf("validator set: %v\n", validatorSet)
		}
		t.Logf("len 2: %v, %v", len(validatorSet.Validators), cap(validatorSet.Validators))
	}
}
