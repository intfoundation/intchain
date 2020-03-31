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

	"enode://aaa40f05d4cc2dedc2536a76ba620d74e533f29991b90490c149bcfb9550c1e0c9bfe63c60cb1f251c06c04d020072c679ef699254b331c59c30c1460bdddabd@127.0.0.1:8551", // test6

}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INTChain test network.
var TestnetBootnodes = []string{
	"enode://bdc5f27a942a3e1f68ab765ba9edc6471eb953e652051b789cb4ee251455b7badf0fa15ed6e14eb79a73d007963b2b578811ba9d319f5e4ad536bf43c6e2472d@129.226.134.100:8550",
}
