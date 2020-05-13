// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the INTChain main network.
var MainnetBootnodes = []string{

	"enode://b8a2baa13828d3a577a3f263374dfecd72a1c12ee07e7828165f33da40e29f29649cd79c10cf63df58c737ee7e2e48231245baa150281481be9c78897029b04c@127.0.0.1:8550", // test6

}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INTChain test network.
var TestnetBootnodes = []string{
	"enode://bdc5f27a942a3e1f68ab765ba9edc6471eb953e652051b789cb4ee251455b7badf0fa15ed6e14eb79a73d007963b2b578811ba9d319f5e4ad536bf43c6e2472d@129.226.134.100:8550",
}
