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
	"enode://9f719ebb3b3f2bf367133b037260e17fad9b81f1ffcf3eadccce3c463444556c09f8b8fcf40c4bcab088220df9d4f15402d0aab38e46c2c4b7f66ec9c34e99a1@129.226.134.100:8550", // Titans
	"enode://131329dda4a526571d343af0d998343914aec081bf4cef4daabaac62c707899fda2a3efaa181640d821053f0d73335fe583826c69e2808961dc323c8d136a1a7@129.226.59.148:8551",  // Oceanus
	"enode://09de3cb98e771c5f2923218d6afa307da3f90ce0a12e0bc33a3a26cb22ab853f04813bdf70d5f6ae3fa8e5b42eb75c314d0fa6572c7c1e85c4efaa62efbc369d@129.226.128.55:8551",  // Iapetus
	"enode://5fe24ed934ab69e38477e89166bc7b98d2ac7067c1b15ac13deba815195fc20722362386bcea5dd8f2918d9a0dd883f7fff62e751f10cd237635b5bec7ae0fb5@129.226.63.13:8551",   // Mnemosyne
}
