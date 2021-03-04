package state

// ----- forbidden Set

// MarkAddressForbidden adds the specified object to the dirty map
//func (self *StateDB) MarkAddressForbidden(addr common.Address) {
//	if _, exist := self.GetForbiddenSet()[addr]; !exist {
//		self.forbiddenSet[addr] = struct{}{}
//		self.forbiddenSetDirty = true
//	}
//}
//
//func (self *StateDB) GetForbiddenSet() ForbiddenSet {
//	if len(self.forbiddenSet) != 0 {
//		return self.forbiddenSet
//	}
//	// Try to get from Trie
//	enc, err := self.trie.TryGet(forbiddenSetKey)
//	if err != nil {
//		self.setError(err)
//		return nil
//	}
//	var value ForbiddenSet
//	if len(enc) > 0 {
//		err := rlp.DecodeBytes(enc, &value)
//		if err != nil {
//			self.setError(err)
//		}
//		self.forbiddenSet = value
//	}
//	return value
//}
//
//func (self *StateDB) commitForbiddenSet() {
//	data, err := rlp.EncodeToBytes(self.forbiddenSet)
//	if err != nil {
//		panic(fmt.Errorf("can't encode forbidden set : %v", err))
//	}
//	self.setError(self.trie.TryUpdate(forbiddenSetKey, data))
//}
//
//func (self *StateDB) ClearForbiddenSetByAddress(addr common.Address) {
//	delete(self.forbiddenSet, addr)
//	self.forbiddenSetDirty = true
//}
//
//// Store the Forbidden Address Set
//
//var forbiddenSetKey = []byte("ForbiddenSet")
//
//type ForbiddenSet map[common.Address]struct{}
//
//func (set ForbiddenSet) EncodeRLP(w io.Writer) error {
//	var list []common.Address
//	for addr := range set {
//		list = append(list, addr)
//	}
//	sort.Slice(list, func(i, j int) bool {
//		return bytes.Compare(list[i].Bytes(), list[j].Bytes()) == 1
//	})
//	return rlp.Encode(w, list)
//}
//
//func (set *ForbiddenSet) DecodeRLP(s *rlp.Stream) error {
//	var list []common.Address
//	if err := s.Decode(&list); err != nil {
//		return err
//	}
//	forbiddenSet := make(ForbiddenSet, len(list))
//	for _, addr := range list {
//		forbiddenSet[addr] = struct{}{}
//	}
//	*set = forbiddenSet
//	return nil
//}
