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

	"enode://c5f758214e3aa79bd077005422e6e5b4abd67c76c298e5c3d0bd15b0d3b80c76eff3f029c2362f4a0cbd4f7b3ef020ca31445c10cb12708b2fee66e3ef698e05@127.0.0.1:8550", // test6

}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://524c80289eb3c87390a145eaf525c23d3063fd73aff2e331f40fcbe7a2b3860dcb2c109d532ee0439780a37f7bf1dc610c64d899e94a9880a0213b935b060d93@129.226.134.100:8550",
	//"enode://22ccd428fd20db673721a09b91984206d52925d2f03338c6dc58af289a0645c7026a2d0098ed4142b6e6ab7496a55e5e4ff8ee8a129f46dc2dcab5c648d30024@129.226.59.148:8551",
	//"enode://5da3651113523acce225e9be3bc06bb8bd75f5f898a4d72132637f5381170672f8a65aed506c8ad3e904bf59e4c4d4f714db187ff0853e1260b66880810cd022@127.0.0.1:8550",
}
