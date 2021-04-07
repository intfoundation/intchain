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
	"enode://cb9dbbd541819bb660568f988c10f5102ff6c38bd2611ec948a777dfe120bc6730584c21dcff0803e3332e6b7b272431551815654b9c6769f96ea1601a59b52c@101.32.74.50:8550", // Titans
	//"enode://66c1cda999dfc803bb29e32602f06873c1a0ed9ab83c0cd5d1b48d2cef715be1641f4a8f942320955150c8f230aaf2b8d71765fc5e8bb9952f70499f4d6ae096@129.226.128.55:8551",  // Oceanus
	//"enode://1840219e9f23fa3e8ccf194735643699e522c7879b39fb12be8b7b8a9b88bcf2e7e8dbd66d5256e677323cc12b3c2e455b6241c70d2e1a452744ce5eded2c73b@129.226.63.13:8551",   // Iapetus
	//"enode://a217493e27effef8975094cd9be107a124aef47cc10f9af6e711c959139e2d0bec6107a9c799a8718c3ab6b0fee99a85d245ca90e01c675cd43a18b6023cf546@170.106.160.155:8551", // Mnemosyne
}
