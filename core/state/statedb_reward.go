package state

import (
	"bytes"
	"fmt"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/rlp"
	"github.com/intfoundation/intchain/trie"
	"io"
	"math/big"
	"sort"
)

// ----- RewardBalance (Total)

// GetTotalRewardBalance Retrieve the reward balance from the given address or 0 if object not found
func (self *StateDB) GetTotalRewardBalance(addr common.Address) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.RewardBalance()
	}
	return common.Big0
}

func (self *StateDB) AddRewardBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		// Add amount to Total Reward Balance
		stateObject.AddRewardBalance(amount)
	}
}

func (self *StateDB) SubRewardBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		// Add amount to Total Reward Balance
		stateObject.SubRewardBalance(amount)
	}
}

// ----- AvailableRewardBalance (Total)

// GetTotalAvailableRewardBalance retrieve the available reward balance from the given address or 0 if object not found
func (self *StateDB) GetTotalAvailableRewardBalance(addr common.Address) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.AvailableRewardBalance()
	}
	return common.Big0
}

func (self *StateDB) AddAvailableRewardBalance(addr common.Address, amount *big.Int) {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		stateObject.AddAvailableRewardBalance(amount)
	}
}

func (self *StateDB) SubAvailableRewardBalance(addr common.Address, amount *big.Int) {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		stateObject.SubAvailableRewardBalance(amount)
	}
}

// ----- Reward Trie

func (self *StateDB) GetDelegateRewardAddress(addr common.Address) map[common.Address]struct{} {
	var deleAddr map[common.Address]struct{}
	reward := Reward{}

	so := self.getStateObject(addr)
	if so == nil {
		return deleAddr
	}

	it := trie.NewIterator(so.getRewardTrie(self.db).NodeIterator(nil))
	for it.Next() {
		var key common.Address
		rlp.DecodeBytes(self.trie.GetKey(it.Key), &key)
		deleAddr[key] = struct{}{}
	}

	if len(so.dirtyReward) > len(deleAddr) {
		reward = so.dirtyReward
		for key := range reward {
			deleAddr[key] = struct{}{}
		}
	}

	if len(so.originReward) > len(deleAddr) {
		reward = so.originReward
		for key := range reward {
			deleAddr[key] = struct{}{}
		}
	}

	return deleAddr
}

// GetRewardBalanceByDelegateAddress
func (self *StateDB) GetRewardBalanceByDelegateAddress(addr common.Address, deleAddress common.Address) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		rewardBalance := stateObject.GetDelegateRewardBalance(self.db, deleAddress)
		if rewardBalance == nil {
			return common.Big0
		} else {
			return rewardBalance
		}
	}
	return common.Big0
}

// AddRewardBalanceByDelegateAddress adds reward amount to the account associated with delegate address
func (self *StateDB) AddRewardBalanceByDelegateAddress(addr common.Address, deleAddress common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		// Get EpochRewardBalance and update EpochRewardBalance
		rewardBalance := stateObject.GetDelegateRewardBalance(self.db, deleAddress)
		var dirtyRewardBalance *big.Int
		if rewardBalance == nil {
			dirtyRewardBalance = amount
		} else {
			dirtyRewardBalance = new(big.Int).Add(rewardBalance, amount)
		}
		stateObject.SetDelegateRewardBalance(self.db, deleAddress, dirtyRewardBalance)

		// Add amount to Total Reward Balance
		stateObject.AddRewardBalance(amount)
	}
}

// AddRewardBalanceByDelegateAddress subtracts reward amount from the account associated with delegate address
func (self *StateDB) SubRewardBalanceByDelegateAddress(addr common.Address, deleAddress common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		// Get EpochRewardBalance and update EpochRewardBalance
		rewardBalance := stateObject.GetDelegateRewardBalance(self.db, deleAddress)
		var dirtyRewardBalance *big.Int
		if rewardBalance == nil {
			panic("you can't subtract the amount from nil balance, check the code, this should not happen")
		} else {
			dirtyRewardBalance = new(big.Int).Sub(rewardBalance, amount)
		}
		stateObject.SetDelegateRewardBalance(self.db, deleAddress, dirtyRewardBalance)

		// Sub amount from Total Reward Balance
		stateObject.SubRewardBalance(amount)
	}
}

func (db *StateDB) ForEachReward(addr common.Address, cb func(key common.Address, rewardBalance *big.Int) bool) {
	so := db.getStateObject(addr)
	if so == nil {
		return
	}
	it := trie.NewIterator(so.getRewardTrie(db.db).NodeIterator(nil))
	for it.Next() {
		var key common.Address
		rlp.DecodeBytes(db.trie.GetKey(it.Key), &key)
		if value, dirty := so.dirtyReward[key]; dirty {
			cb(key, value)
			continue
		}
		var value big.Int
		rlp.DecodeBytes(it.Value, &value)
		cb(key, &value)
	}
}

// ----- Reward Set

// MarkAddressReward adds the specified object to the dirty map to avoid
func (self *StateDB) MarkAddressReward(addr common.Address) {
	if _, exist := self.GetRewardSet()[addr]; !exist {
		self.rewardSet[addr] = struct{}{}
		self.rewardSetDirty = true
	}
}

func (self *StateDB) GetRewardSet() RewardSet {
	if len(self.rewardSet) != 0 {
		return self.rewardSet
	}
	// Try to get from Trie
	enc, err := self.trie.TryGet(rewardSetKey)
	if err != nil {
		self.setError(err)
		return nil
	}
	var value RewardSet
	if len(enc) > 0 {
		err := rlp.DecodeBytes(enc, &value)
		if err != nil {
			self.setError(err)
		}
		self.rewardSet = value
	}
	return value
}

func (self *StateDB) commitRewardSet() {
	data, err := rlp.EncodeToBytes(self.rewardSet)
	if err != nil {
		panic(fmt.Errorf("can't encode reward set : %v", err))
	}
	self.setError(self.trie.TryUpdate(rewardSetKey, data))
}

func (self *StateDB) ClearRewardSetByAddress(addr common.Address) {
	delete(self.rewardSet, addr)
	self.rewardSetDirty = true
}

// Store the Reward Address Set

var rewardSetKey = []byte("RewardSet")

type RewardSet map[common.Address]struct{}

func (set RewardSet) EncodeRLP(w io.Writer) error {
	var list []common.Address
	for addr := range set {
		list = append(list, addr)
	}
	sort.Slice(list, func(i, j int) bool {
		return bytes.Compare(list[i].Bytes(), list[j].Bytes()) == 1
	})
	return rlp.Encode(w, list)
}

func (set *RewardSet) DecodeRLP(s *rlp.Stream) error {
	var list []common.Address
	if err := s.Decode(&list); err != nil {
		return err
	}
	rewardSet := make(RewardSet, len(list))
	for _, addr := range list {
		rewardSet[addr] = struct{}{}
	}
	*set = rewardSet
	return nil
}

// ----- Child Chain Reward Per Block

func (self *StateDB) SetChildChainRewardPerBlock(rewardPerBlock *big.Int) {
	self.childChainRewardPerBlock = rewardPerBlock
	self.childChainRewardPerBlockDirty = true
}

func (self *StateDB) GetChildChainRewardPerBlock() *big.Int {
	if self.childChainRewardPerBlock != nil {
		return self.childChainRewardPerBlock
	}
	// Try to get from Trie
	enc, err := self.trie.TryGet(childChainRewardPerBlockKey)
	if err != nil {
		self.setError(err)
		return nil
	}
	value := new(big.Int)
	if len(enc) > 0 {
		err := rlp.DecodeBytes(enc, value)
		if err != nil {
			self.setError(err)
		}
		self.childChainRewardPerBlock = value
	}
	return value
}

func (self *StateDB) commitChildChainRewardPerBlock() {
	data, err := rlp.EncodeToBytes(self.childChainRewardPerBlock)
	if err != nil {
		panic(fmt.Errorf("can't encode child chain reward per block : %v", err))
	}
	self.setError(self.trie.TryUpdate(childChainRewardPerBlockKey, data))
}

// Child Chain Reward Per Block

var childChainRewardPerBlockKey = []byte("RewardPerBlock")

func (self *StateDB) MarkProposedInEpoch(address common.Address, epoch uint64) error {

	return self.db.TrieDB().MarkProposedInEpoch(address, epoch)
}

func (self *StateDB) CheckProposedInEpoch(address common.Address, epoch uint64) bool {

	return self.db.TrieDB().CheckProposedInEpoch(address, epoch)
}
