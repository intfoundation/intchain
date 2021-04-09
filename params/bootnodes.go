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
	"enode://982872e8d8d97b66151d83a334d80bbbedac739a987fc5a7c472c230c25576969d3488a271c6f355606973832461c773d00e7ade9e9ccc8b0ca46b10a1cea1a1@101.32.74.50:8550",    // Titans
	"enode://9381a60ca2c65cb649c56d84789bb1164cafd766a754ed484e414f5b02bc6a622bc8d360043fb0e776471119229fff74d358f57415730d7f3d62d3c0155666d5@129.226.128.55:8551",  // Oceanus
	"enode://b82aa7354bcf98cfe4eb07da7bb39b22ae27618165ab9136a3a6525c3ab7b87114fc7743b4dff312acc1eb31185a22d2bb097fd0692f6bf3ccaa98b605b140d0@129.226.63.13:8551",   // Iapetus
	"enode://99a74bc83ea34e24bc7b6fa6dcf8e1f533febb296c7000f4b9f1f873365ba6514435d35fe0182014bac31b3c0f4e21398f398cfe089c6b284f4c5fe2a5e5acd3@170.106.160.155:8551", // Mnemosyne
}
