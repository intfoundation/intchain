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

	"enode://7fc4452e7199520ff8db7727604ac194a47ec6a75e3a2e7e9451853b5bf915a3dbfd1958c9116c3d4b151e6e5b55e340eb0b6ce517aff0b9a2ce630b514c3a74@127.0.0.1:8550",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://20c9a43afb9b19ada68883ef404810d93963be0e09f90444cce52e9a68fbdb9b7fec0187dd1e2c31737c0b452884a43e0fd654bde18310588b368d080836f8f9@101.32.74.50:8550",   // Titans
	"enode://244f36c3c07992906ce71f01da98453f0ec1218eae1895b6c5382052f3e8cb7bebb3b910eec39b94673a2358f8d6ebf91bf7c4e3bd9553da9f7953a67cd8da9c@129.226.59.148:8551", // Oceanus
	"enode://0ed1e56f70582cdea3c7bf1004b3af7a6e8f395958b5ed71efb9201e1340e6ab5bf3a7d0f3250dcaf92fd79cf00e957ebf4738c87ad696cd09290a6b880acf13@129.226.128.55:8551", // Iapetus
	"enode://febc48785b735f5ac4a50e9d5265748c66ee6d767267c84ceb4ebc8a1ba7ff41f622fa55a3ded6b7bb1c60189e0a9ebc0e61d973cda434057431d9e239685e99@129.226.63.13:8551",  // Mnemosyne
}
