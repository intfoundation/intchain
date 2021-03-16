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

	"enode://b49029664a737763951a65acc5239ef510ff47ccceb067a0b72253c6e744d5cacc51b129c852fd980b72b64606aa5b6caef6475bee5dc812ef48210d63604c8b@127.0.0.1:8550",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://686a71fadce77a27610a7ecab1f974da0c90a237df427189e63794405cb6d367c2a85aba69bf626d00fffb87a0c80a2aa115e21a3d9aef360c2cdc94948e62ec@101.32.74.50:8550",    // Titans
	"enode://728a6c64dc7a21d138ad5c0ee30ed09a891a0f399abd3cf5801a460fa25b28e96d7f630c3e26893e0448dad6802b036f1b4f47008d8a377064e2d980668473ab@129.226.128.55:8551",  // Oceanus
	"enode://4cd0e4c1e09b13ccbe4430b2d5d55dc41305f052770606e3eda04851434a70be2f2edb9644f42e9c62bc0c5d1588e27057cde5756f0a76f4f58b552a16ffc99d@129.226.63.13:8551",   // Iapetus
	"enode://55a833750b44fe818df2bc008af225cef4fb7a13b21b4420381db82b5efc5bc5437b4d05beb8145aadb7de257e589016a17c415c1f5263c8816d44387a1417fc@170.106.160.155:8551", // Mnemosyne
}
