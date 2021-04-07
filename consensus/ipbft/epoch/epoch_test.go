package epoch

import (
	"fmt"
	"github.com/intfoundation/intchain/common"
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
