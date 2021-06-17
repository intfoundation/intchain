package types

import (
	"bytes"
	"fmt"
	"github.com/intfoundation/go-wire"
)

// Information for each candidate
type Candidate struct {
	Address []byte `json:"address"`
}

func NewCandidate(address []byte) *Candidate {
	return &Candidate{
		Address: address,
	}
}

func (c *Candidate) Copy() *Candidate {
	cCopy := *c
	return &cCopy
}

func (c *Candidate) Equals(other *Candidate) bool {
	return bytes.Equal(c.Address, other.Address)
}

func (c *Candidate) String() string {
	if c == nil {
		return "nil-Candidate"
	}

	return fmt.Sprintf("Candidate{ADD:%X}",
		c.Address)
}

func (c *Candidate) Hash() []byte {
	return wire.BinaryRipemd160(c)
}
