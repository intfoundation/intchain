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
	"enode://3dceb610c7c33ab136557a4bb92f58e569d8d582c2c3e1256842bbb7de0237902c9edbaad85f92b906b27ec2895d6b73be9613fb5891e7d2b4124139698521f3@129.226.134.100:8550", // Titans
	"enode://1a13ee4c909ed3a433454adbe70acd2a8147c68799b461d5e9c464c1b8a8caec7265f6d33de504977483d66cb7cce3c1c9fe8b0a2c44608463be138c64ca74f9@129.226.59.148:8551",  // Oceanus
	"enode://ac9555187e3173340f4143903a9a00b19e3aa0e6193e7141c40a476844242de179428fbbae9899ec0838e8ca74f0255c45ef4e08a69ac7f75d5cf4e1ab566942@129.226.128.55:8551",  // Iapetus
	"enode://00efe86d3ec4e9163cfba375bcffd8f93058f15a05896b4846157de09353f55936eb3fa36cebea17835583cf775041a46ae7cd2a2989bcd1130eaccaa5fb3b09@129.226.63.13:8551",   // Mnemosyne
}
