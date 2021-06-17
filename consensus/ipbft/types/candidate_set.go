package types

import (
	"bytes"
	"sort"
)

type CandidateSet struct {
	Candidates []*Candidate `json:"candidates"`
}

func NewCandidateSet(cans []*Candidate) *CandidateSet {
	candidates := make([]*Candidate, len(cans))
	for i, can := range cans {
		candidates[i] = can.Copy()
	}

	sort.Sort(CandidatesByAddress(candidates))

	cs := &CandidateSet{
		Candidates: candidates,
	}
	return cs
}

func (canSet *CandidateSet) Copy() *CandidateSet {
	candidates := make([]*Candidate, len(canSet.Candidates))
	for i, can := range canSet.Candidates {
		candidates[i] = can.Copy()
	}

	return &CandidateSet{
		Candidates: candidates,
	}
}

func (canSet *CandidateSet) Add(cal *Candidate) (added bool) {
	cal = cal.Copy()

	idx := -1
	for i := 0; i < len(canSet.Candidates); i++ {
		if bytes.Compare(cal.Address, canSet.Candidates[i].Address) == 0 {
			idx = i
			break
		}
	}

	//if idx == len(valSet.Validators) {
	if idx == -1 {
		canSet.Candidates = append(canSet.Candidates, cal)
		return true
	} else {
		return false
	}
}

func (canSet *CandidateSet) Remove(address []byte) (cal *Candidate, removed bool) {
	idx := -1
	for i := 0; i < len(canSet.Candidates); i++ {
		if bytes.Compare(address, canSet.Candidates[i].Address) == 0 {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, false
	} else {
		removedCal := canSet.Candidates[idx]
		newCandidates := canSet.Candidates[:idx]
		if idx+1 < len(canSet.Candidates) {
			newCandidates = append(newCandidates, canSet.Candidates[idx+1:]...)
		}
		canSet.Candidates = newCandidates
		return removedCal, true
	}
}

// HasAddress returns true if address given is in the candidate set, false -
// otherwise.
func (canSet *CandidateSet) HasAddress(address []byte) bool {

	for i := 0; i < len(canSet.Candidates); i++ {
		if bytes.Compare(address, canSet.Candidates[i].Address) == 0 {
			return true
		}
	}
	return false
}

func (canSet *CandidateSet) GetByAddress(address []byte) (index int, val *Candidate) {

	idx := -1
	for i := 0; i < len(canSet.Candidates); i++ {
		if bytes.Compare(address, canSet.Candidates[i].Address) == 0 {
			idx = i
			break
		}
	}

	if idx != -1 {
		return idx, canSet.Candidates[idx].Copy()
	} else {
		return 0, nil
	}
}

type CandidatesByAddress []*Candidate

func (cs CandidatesByAddress) Len() int {
	return len(cs)
}

func (cs CandidatesByAddress) Less(i, j int) bool {
	return bytes.Compare(cs[i].Address, cs[j].Address) == -1
}

func (cs CandidatesByAddress) Swap(i, j int) {
	it := cs[i]
	cs[i] = cs[j]
	cs[j] = it
}
