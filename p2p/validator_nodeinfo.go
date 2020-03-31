package p2p

import (
	"time"

	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/p2p/discover"
	"github.com/intfoundation/intchain/rlp"
	"golang.org/x/crypto/sha3"
)

type P2PValidator struct {
	ChainId string
	Address common.Address
}

type P2PValidatorNodeInfo struct {
	Node      discover.Node
	TimeStamp time.Time
	Validator P2PValidator
	Original  bool
}

func (vni *P2PValidatorNodeInfo) Hash() common.Hash {
	return rlpHash(vni)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
