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
	"enode://8d92a0dfb34ef659fdd6b3a157b753d656986e937bd2b8ea91fb685f2b1c8ab51299412d561842d8ea3a20eeb834312fd396c0d648780b7475a7202878db6829@159.138.3.248:8551",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://dd7e1ead96067f73ed13c5a7a0e6d1cf7a92644a225478acda42c48750aff021263740481b5d1d72275fe19e4d61d064bbb3293adba0f0845d013f7ece63c75d@101.32.74.50:8550",   // Titans
	"enode://4dd8d8352900f44cc6bfd9d9c3f2d9c8d80d0b6a64b97019dc9e3ca905a4ca6474b84745e5001fc08c5b64910baeb233fcd54bda4f6f855438d7c762494977c0@129.226.59.148:8551", // Oceanus
	"enode://964f2c8d04bcc48e971bb24870035db8cfc7218bfdb4d97849421b47d9cc207fff1b89876b4ff137156fb52a262254a570396a4c06a756d2c68a8f637eaa5031@129.226.128.55:8551", // Iapetus
	"enode://65e60686f0c091b30a701b3c8e02db1aab75ad67e7d319c679a616cad38ce2dad5abb8829cd0d6041a9541387078eed12b88f5d0eb5168456a182ad2d0183e19@129.226.63.13:8551",  // Mnemosyne
}
