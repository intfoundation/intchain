package consensus

import (
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/consensus/ipbft/types"
	"testing"
)

var (
	testProposal types.Proposal
	testPubKey   crypto.BLSPubKey
)

func TestVerifyBytes(t *testing.T) {
	testPubKeyBytes, err := hexutil.Decode("0x6972968EA971CEBDFFD9DA316C5233060B09141EF69573DDBD0CD503C7179369220B750C50F8867B37FD9341C73C90C650911B2BA849A94E253A0B09461632F68F9AD8CF3867B5CF3F7A0A543A23E22239B7628A2EEE5E0F7F8B1309B668FAFA754430E6CDF22758452979E15912A1B5162AFE918DAAE73EAD75C32AD42F66CD")
	if err != nil {
		t.Errorf("decode public key error %v\n", err)
	}

	copy(testPubKey[:], testPubKeyBytes)

	//fmt.Printf("pubkey %v\n", testPubKey)

	//testProposal = types.NewProposal(uint64(675224), int(4))
}
