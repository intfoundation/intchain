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
// the INT Chain main network.
var MainnetBootnodes = []string{

	"enode://bbe4ef6582707fd70d98c4cd35a308883d6f920b72a6caf5c8c3c42c3a2455be893759baf2f3fbbfae972247242255b2b25bc033267b3d680877a19b8d36e1af@127.0.0.1:8550",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://686a71fadce77a27610a7ecab1f974da0c90a237df427189e63794405cb6d367c2a85aba69bf626d00fffb87a0c80a2aa115e21a3d9aef360c2cdc94948e62ec@101.32.74.50:8550", // Titans
}
