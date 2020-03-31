package types

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/intfoundation/bls"
	"github.com/intfoundation/intchain/common"
	. "github.com/intfoundation/go-common"
	"github.com/intfoundation/go-crypto"
	"github.com/intfoundation/go-wire"
)

type PrivValidator struct {
	// IntChain Account Address
	Address common.Address `json:"address"`
	// IntChain Consensus Public Key, in BLS format
	PubKey crypto.PubKey `json:"consensus_pub_key"`
	// IntChain Consensus Private Key, in BLS format
	// PrivKey should be empty if a Signer other than the default is being used.
	PrivKey crypto.PrivKey `json:"consensus_priv_key"`

	Signer `json:"-"`

	// For persistence.
	// Overloaded for testing.
	filePath string
	mtx      sync.Mutex
}

// 修改 privalidator 的地址类型为string
type PrivV struct {
	Address string         `json:"address"`
	PubKey  crypto.PubKey  `json:"consensus_pub_key"`
	PrivKey crypto.PrivKey `json:"consensus_priv_key"`

	Signer `json:"-"`

	// For persistence.
	// Overloaded for testing.
	filePath string
	mtx      sync.Mutex
}

// This is used to sign votes.
// It is the caller's duty to verify the msg before calling Sign,
// eg. to avoid double signing.
// Currently, the only callers are SignVote and SignProposal
type Signer interface {
	Sign(msg []byte) crypto.Signature
}

// Implements Signer
type DefaultSigner struct {
	priv crypto.PrivKey
}

func NewDefaultSigner(priv crypto.PrivKey) *DefaultSigner {
	return &DefaultSigner{priv: priv}
}

// Implements Signer
func (ds *DefaultSigner) Sign(msg []byte) crypto.Signature {
	return ds.priv.Sign(msg)
}

func GenPrivValidatorKey(address common.Address) *PrivValidator {

	keyPair := bls.GenerateKey()
	var blsPrivKey crypto.BLSPrivKey
	copy(blsPrivKey[:], keyPair.Private().Marshal())

	blsPubKey := blsPrivKey.PubKey()

	return &PrivValidator{
		Address: address,
		PubKey:  blsPubKey,
		PrivKey: blsPrivKey,

		filePath: "",
		Signer:   NewDefaultSigner(blsPrivKey),
	}
}

//func LoadPrivValidator(filePath string) *PrivValidator {
//	privValJSONBytes, err := ioutil.ReadFile(filePath)
//	if err != nil {
//		Exit(err.Error())
//	}
//	privVal := wire.ReadJSON(&PrivValidator{}, privValJSONBytes, &err).(*PrivValidator)
//	if err != nil {
//		Exit(Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
//	}
//	privVal.filePath = filePath
//	privVal.Signer = NewDefaultSigner(privVal.PrivKey)
//	return privVal
//}

// 修改加载文件的方法，兼容string 的地址类型
func LoadPrivValidator(filePath string) *PrivValidator {
	privValJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		Exit(err.Error())
	}
	privVal := wire.ReadJSON(&PrivV{}, privValJSONBytes, &err).(*PrivV)

	if err != nil {
		Exit(Fmt("Error reading PrivValidator from %v: %v\n", filePath, err))
	}
	privV := &PrivValidator{
		Address:  common.StringToAddress(privVal.Address),
		PubKey:   privVal.PubKey,
		PrivKey:  privVal.PrivKey,
		filePath: filePath,
		Signer:   NewDefaultSigner(privVal.PrivKey),
	}

	return privV
}

func (pv *PrivValidator) SetFile(filePath string) {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	pv.filePath = filePath
}

func (pv *PrivValidator) Save() {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	pv.save()
}

//func (pv *PrivValidator) save() {
//	if pv.filePath == "" {
//		PanicSanity("Cannot save PrivValidator: filePath not set")
//	}
//	jsonBytes := wire.JSONBytesPretty(pv)
//	// 使用 WriteFileAtomic（）方法，文件里面的地址为16进制的
//	err := WriteFileAtomic(pv.filePath, jsonBytes, 0600)
//	if err != nil {
//		// `@; BOOM!!!
//		PanicCrisis(err)
//	}
//}

func (pv *PrivValidator) save() {
	if pv.filePath == "" {
		PanicSanity("Cannot save PrivValidator: filePath not set")
	}
	//jsonBytes := wire.JSONBytesPretty(pv)
	var priv PrivV
	priv.Address = pv.Address.String()
	priv.PubKey = pv.PubKey
	priv.PrivKey = pv.PrivKey

	jsonBytes := wire.JSONBytesPretty(priv)
	err := WriteFileAtomic(pv.filePath, jsonBytes, 0600)
	if err != nil {
		// `@; BOOM!!!
		PanicCrisis(err)
	}
}

func (pv *PrivValidator) GetAddress() []byte {
	return pv.Address.Bytes()
}

func (pv *PrivValidator) GetPubKey() crypto.PubKey {
	return pv.PubKey
}

func (pv *PrivValidator) SignVote(chainID string, vote *Vote) error {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	signature := pv.Sign(SignBytes(chainID, vote))
	vote.Signature = signature
	return nil
}

func (pv *PrivValidator) SignProposal(chainID string, proposal *Proposal) error {
	pv.mtx.Lock()
	defer pv.mtx.Unlock()

	signature := pv.Sign(SignBytes(chainID, proposal))
	proposal.Signature = signature
	return nil
}

func (pv *PrivValidator) String() string {
	return fmt.Sprintf("PrivValidator{%X}", pv.Address)
}

//-------------------------------------

type PrivValidatorsByAddress []*PrivValidator

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(pvs[i].Address[:], pvs[j].Address[:]) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	it := pvs[i]
	pvs[i] = pvs[j]
	pvs[j] = it
}
