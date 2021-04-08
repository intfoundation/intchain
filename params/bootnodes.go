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
	"enode://cb9dbbd541819bb660568f988c10f5102ff6c38bd2611ec948a777dfe120bc6730584c21dcff0803e3332e6b7b272431551815654b9c6769f96ea1601a59b52c@101.32.74.50:8550",    // Titans
	"enode://b6e88380e806a2e36af597578213476a80636f802fc6be5b77eb8ecff83b51e3984c10987560c34fd51bbe713fff63e7d8b5b81292ac4ec61f516a7d2fd7c96b@129.226.128.55:8551",  // Oceanus
	"enode://a7a6190dd0642598306d6e0f537d7594dc9115d0eaf302b50f7fb7bb2f1c9d00350eaeec575570f29c92cc3277847c69a7df621ee23b55b00ebeab04b59602df@129.226.63.13:8551",   // Iapetus
	"enode://460d595d4b0ba6466a0b254b4cb6cb21b76cfd0b9b6ac86202915bb36e090c529b9351410aa57c7fd9d88a1710295ebd9a1aedf7f57be517056bf38f3296a5c7@170.106.160.155:8551", // Mnemosyne
}
