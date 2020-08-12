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

	"enode://c5f758214e3aa79bd077005422e6e5b4abd67c76c298e5c3d0bd15b0d3b80c76eff3f029c2362f4a0cbd4f7b3ef020ca31445c10cb12708b2fee66e3ef698e05@127.0.0.1:8550",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://7b7448e679a7ada82420426a59e9aef057b81c3e931199e83de0b237e7e47f1a3fab2c964d4d8450785905b6a34a2049963eda87c426f69cf74463b870e5e3c3@129.226.134.100:8550", // Titans
}
