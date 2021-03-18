package intapi

import (
	"bytes"
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/crypto"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"testing"
	"time"
)

type MethoadParams struct {
	Input   string
	Args    interface{}
	FunType intAbi.FunctionType
}

var inputHex = "0x91e8537e000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000044c696b6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001068747470733a2f2f6c696b652e636f6d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000104531353733453236384138313835303300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c4d792076616c696461746f720000000000000000000000000000000000000000"

var inputArray = []*MethoadParams{
	{
		Input:   "0x49339f0f00000000000000000000000026ee0906f135303a0ab66b3196efabd0853c481b",
		Args:    intAbi.DelegateArgs{},
		FunType: intAbi.Delegate,
	},
	{
		Input:   "0x91e8537e000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000044c696b6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001068747470733a2f2f6c696b652e636f6d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000104531353733453236384138313835303300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c4d792076616c696461746f720000000000000000000000000000000000000000",
		Args:    intAbi.EditValidatorArgs{},
		FunType: intAbi.EditValidator,
	},
}

func TestMethodId(t *testing.T) {
	method := intAbi.ChainABI.Methods[intAbi.Delegate.String()]
	methdid := method.ID()

	fmt.Printf("method id, %v", hexutil.Encode(methdid))
}

func TestABI_UnpackMethodInputs(t *testing.T) {

	//inputByte, err := hexutil.Decode(inputHex)
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//function, err := intAbi.FunctionTypeFromId(inputByte[:4])
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Printf("function=%v\n", function)
	//
	////var args intAbi.CandidateArgs
	//var args intAbi.DelegateArgs
	////var args intAbi.EditValidatorArgs
	//
	//err = intAbi.ChainABI.UnpackMethodInputs(&args, function.String(), inputByte[4:])
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Printf("unpack method %v\n", args.Moniker)
	//fmt.Printf("unpack website %v\n", args.Website)
	//fmt.Printf("unpack identify %v\n", args.Identity)
	//fmt.Printf("unpack details %v\n", args.Details)

	for _, v := range inputArray {
		inputByte, err := hexutil.Decode(v.Input)
		if err != nil {
			t.Error(err)
		}

		err = checkFunType(inputByte, v.FunType)
		if err != nil {
			t.Error(err)
		}

		unpackArgs, err := unpackMethod(inputByte, v.FunType)
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("unpack %v success,  args %v", v.FunType.String(), unpackArgs.Candidate.String())
		}

	}
}

func checkFunType(input []byte, funType intAbi.FunctionType) error {
	function, err := intAbi.FunctionTypeFromId(input[:4])

	if err != nil {
		return err
	}

	if !bytes.Equal([]byte(function.String()), []byte(funType.String())) {
		return fmt.Errorf("method mismatch want %v, but %v", funType.String(), function.String())
	}

	return nil
}

func unpackMethod(input []byte, funType intAbi.FunctionType) (unpackArgs intAbi.DelegateArgs, err error) {
	var args intAbi.DelegateArgs

	err = intAbi.ChainABI.UnpackMethodInputs(&args, funType.String(), input[4:])
	if err != nil {
		return unpackArgs, err
	}

	return args, nil

}

var FromAddr = "INT3MccCA7EtMzijJa2zjxoiSYzbNLE4"
var PubKey = "0x618CEAF6AD449B826E2521222A94426B82800202332251F0929EC47B36A647C65E00D2EA34C07A8EF7953C2E1555D8321449423CCFB0B64BB13090E7A433114D68F1C1891BAA20101E5CC8E2B10E207F5D21D1A1116547E1EED5E92FDFE4F5E58119C5267B82AE06BBA5016827396B74E1ECDCC3801746242CA24C7749EB2F88"
var Amount = "0x152d02c7e14af68000000"
var Salt = "like"
var VoteHash1 = "0xc6335e23dd8ba330b2d3c34acdeb2dfd0b07d30dfc2d5f9ca1b0d62e147788f0" // false
var VoteHash2 = "0xa431ab9cb5d2750faeed74945d10c69372b938c2470d5b140de29f4d4aa22025" // true
var VoteHash3 = "0xb2aa67b3cf56dcb41097d72024962c03d4fba2a9892cc37e348243b85bf58c27" // false

func TestVoteHash(t *testing.T) {
	pubKey, err := hexutil.Decode(PubKey)
	if err != nil {
		t.Fatal(err)
	}
	byteData := [][]byte{
		[]byte(FromAddr),
		pubKey,
		common.LeftPadBytes(math.MustParseBig256("1600000000000000000000000").Bytes(), 1),
		[]byte(Salt),
	}

	hash := crypto.Keccak256Hash(concatCopyPreAllocate(byteData))
	fmt.Printf("vote hash %v\n", hash.String())

}

func TestGoTime(t *testing.T) {
	nowTime := time.Now().Unix()
	fmt.Printf("now %v\n", nowTime)

	d := 24 * time.Hour
	fmt.Printf("duration %v\n", d)
	fmt.Printf("duration string %v\n", d.String())
	fmt.Printf("duration seconds %v\n", d.Seconds())
}
