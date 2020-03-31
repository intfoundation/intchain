package ethapi

import (
	"bytes"
	"fmt"
	"github.com/intfoundation/intchain/common/hexutil"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"testing"
)

type MethoadParams struct {
	Input   string
	Args    interface{}
	FunType intAbi.FunctionType
}

var inputHex = "0x91e8537e000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000044c696b6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001068747470733a2f2f6c696b652e636f6d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000104531353733453236384138313835303300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c4d792076616c696461746f720000000000000000000000000000000000000000"

var inputArray = []*MethoadParams{
	{
		Input:   "0x49339f0f494e5433437046756b32634a31746539575a563177385933776b51436341355a",
		Args:    intAbi.DelegateArgs{},
		FunType: intAbi.Delegate,
	},
	{
		Input:   "0x91e8537e000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000014000000000000000000000000000000000000000000000000000000000000000044c696b6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001068747470733a2f2f6c696b652e636f6d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000104531353733453236384138313835303300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c4d792076616c696461746f720000000000000000000000000000000000000000",
		Args:    intAbi.EditValidatorArgs{},
		FunType: intAbi.EditValidator,
	},
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
