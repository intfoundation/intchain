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

	"enode://b8a2baa13828d3a577a3f263374dfecd72a1c12ee07e7828165f33da40e29f29649cd79c10cf63df58c737ee7e2e48231245baa150281481be9c78897029b04c@127.0.0.1:8550", // test6

}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://2efb6005361c4347f2fe703f34ba24206254cb07f18ba827c6e295c4508cebdfb127f1b124dc7cadacc987155bedab87404f48926970a1430987b14060e5896f@129.226.134.100:8550",
	//"enode://22ccd428fd20db673721a09b91984206d52925d2f03338c6dc58af289a0645c7026a2d0098ed4142b6e6ab7496a55e5e4ff8ee8a129f46dc2dcab5c648d30024@129.226.59.148:8551",
	//"enode://5da3651113523acce225e9be3bc06bb8bd75f5f898a4d72132637f5381170672f8a65aed506c8ad3e904bf59e4c4d4f714db187ff0853e1260b66880810cd022@127.0.0.1:8550",
}
