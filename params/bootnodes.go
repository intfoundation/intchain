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
	//"enode://6134c7e0034ac4ae16d0f16ba0fb08aa3ebdc5a141c566241ee98a87b2f805e85c8671a03fb916ce445a8938b9f8f5fd48310c6469314a44a80b701bf09e63fe@101.32.74.50:8550",   // Titans
	//"enode://d5f3fc19f459b7684f2f9c27170f719551763bd2d8c1c90cdc6c84fb537deb56866d1bd4a15150e65dced40206c8e4d686058e0251d7db194435c4689ec3ad75@129.226.59.148:8551", // Oceanus
	//"enode://7e6322f103803be6f3f5d8faba8da2effd8462e6171435dbff723d55ba30df61bb0145fa8b826644bc99c0207c8dc198d95fc8cab06ce53c140f0136bf95f9ef@129.226.128.55:8551", // Iapetus
	//"enode://b1b1d48898bfa0c0a40700c351cc50e0a9d2407c7ce952dc4d6e671cbe1921d0ba79fb88acedcd8d633a2a19e6951d0837a99a6a5169a0dbd74122b6584a99e8@129.226.63.13:8551",  // Mnemosyne
	"enode://1a1bc3e3108937947f7375c7fc44559eb4879f8cde95a2d6fd2ccde57897621bd7f644ef94ab5a0c225aa6ba093a1734daee098216191b70bc95755aa524d5e1@192.168.188.110:8550", // Local
	"enode://de11cbd922a4c5ec04ced7ac0633c8d477f10f5e838f96216a6efeaa3d70c90e5d429ed4bb1c4b94d58aea37b98e47ca47e2af87f648604877deff3035c20261@192.168.188.225:8550", // Local
}
