package state

import (
	"bytes"
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/rlp"
	"io"
	"sort"
)

// ----- candidate Set

// MarkAddressCandidate adds the specified object to the dirty map
func (self *StateDB) MarkAddressCandidate(addr common.Address) {
	if _, exist := self.GetCandidateSet()[addr]; !exist {
		self.candidateSet[addr] = struct{}{}
		self.candidateSetDirty = true
	}
}

func (self *StateDB) ClearCandidateSetByAddress(addr common.Address) {
	fmt.Printf("candidate set bug, clear candidate set by address, %v\n", self.candidateSet)
	delete(self.candidateSet, addr)
	self.candidateSetDirty = true
}

func (self *StateDB) GetCandidateSet() CandidateSet {
	fmt.Printf("candidate set bug, get candidate set 1, %v\n", self.candidateSet)
	if len(self.candidateSet) != 0 {
		return self.candidateSet
	}
	// Try to get from Trie
	enc, err := self.trie.TryGet(candidateSetKey)
	if err != nil {
		self.setError(err)
		return nil
	}
	var value CandidateSet
	if len(enc) > 0 {
		err := rlp.DecodeBytes(enc, &value)
		if err != nil {
			self.setError(err)
		}
		self.candidateSet = value
	}
	fmt.Printf("candidate set bug, get candidate set 2, %v\n", value)
	return value
}

func (self *StateDB) commitCandidateSet() {
	//fmt.Printf("candidate set bug, commit candidate set, %v\n", self.candidateSet)
	data, err := rlp.EncodeToBytes(self.candidateSet)
	if err != nil {
		panic(fmt.Errorf("can't encode candidate set : %v", err))
	}
	self.setError(self.trie.TryUpdate(candidateSetKey, data))
}

// Store the Candidate Address Set

var candidateSetKey = []byte("CandidateSet")

type CandidateSet map[common.Address]struct{}

func (set CandidateSet) EncodeRLP(w io.Writer) error {
	var list []common.Address
	for addr := range set {
		list = append(list, addr)
	}
	sort.Slice(list, func(i, j int) bool {
		return bytes.Compare(list[i].Bytes(), list[j].Bytes()) == 1
	})
	return rlp.Encode(w, list)
}

func (set *CandidateSet) DecodeRLP(s *rlp.Stream) error {
	var list []common.Address
	if err := s.Decode(&list); err != nil {
		return err
	}
	candidateSet := make(CandidateSet, len(list))
	for _, addr := range list {
		candidateSet[addr] = struct{}{}
	}
	*set = candidateSet
	return nil
}
