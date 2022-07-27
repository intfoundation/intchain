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
	"enode://191df92ecc4ffd4df73a512ece6f66324361fe7a2efdbcc222a826692e55d4c22c0c7d0b2fefc2f3973eb23f18cd6b3dce3552f331b9848d0c15d66c65648556@43.129.172.110:8551", // INT Chain
	"enode://a201378734ae3a7ae9884c0009fa611752fae87903696fdcd78a2bcff47a83893e8cde2d30c90d602a9c19de93138f9ac06aa9bf8951b68905779b2e8088a60f@43.128.19.174:8552",  // Crius
	"enode://f5783bbeecefe6d80814e8bc928403f0254a90f0e353160cda54c8f7257c10dd13d5e984763f9cd292fb7e9b300bc772b46bb08952d58645b224f8eaf3a861e6@43.129.69.192:8551",  //
	"enode://f27ed1c0a485ab019e026a71162390b7fcf9ac16cc03fe20c3aeb0b214f643a28b79ef44e68321d002f3b51627886681ac0dd8462a145099ba69835d7b666a71@43.129.237.219:8552", // Mnemosyne
	"enode://2b8b2c7584e07bec2ef156775a8c4149162dc7805617314bc795048a117af912458394e4492897d54bfcc8f672f67a2cbac08631e5001722c00dfd98df3f2ace@43.129.237.184:8551", // fengli

	"enode://374f9f738071254049d13e2c50048b09b2f20fa8e25073a89a51e22040a5e2da2f57febfa56618f147cb3cd95a0e0bb0ebd7d00789d7da86b2440efc9d1d7bee@43.129.69.205:8552", // Community
	"enode://a600a79296f0487cc1ed72a46fdb50e48d075a37a28e6597b77d5e60595d55c7d02997fed58ceeb38441130829f20bed925eea060eb51d0390bb45f89a1ca45d@43.129.69.39:8552",  // Community
	"enode://03784c32e5a4e0d1aced08f6090d9159e02b237c34f8a4bd60d2a9427c549b17d718f7a20b16b6a1a2b33aeecca73db37765de61b32d3d8ec4bab8e15c4df904@43.129.66.168:8553", // Community
	"enode://1bc59b88e8a8a1c3960d7f2753a12d86b17e09f15d2419ce72f9923850f0f5bd370ec3730d4c5847e4fa80267c8136ca212f2bb648cc81ff3530f53a0997723f@43.129.69.166:8553", // Community
	"enode://3e421e5d549050ceaca43e7cccb7522a1949b8cf2c598dff36da2bc164c90c8996bb49c442a6e798f7aecbd3ca63f85471c28b958ca78d955a4513ac2373e735@43.132.154.65:8553", // Community

	"enode://9260fcf0ff3996607d9e1ea7b222107ad3eaf657b1c8aa760690bf971f7e4151c6271a2c87412b31312c76b3d0fb285947d3832e6cb972db14a5be9a45cb7873@162.55.177.167:8550",  // Community
	"enode://d71a17716ee2a41c270e01ab63c28863f09fddd1e1e5abbc0d5cdc623e927529b502118812a902ebc868bd76570c851ecb50620cda19333a4ca92a9c32f348c9@49.12.192.182:8550",   // Community
	"enode://ef1979cc634da020a132f6c0ea53cb112f7c1eabc6d285e977bd75eb79dee89693c30686dc52575510e6b1e664427e25ea2c2d36d4630371ae1d250f8ec110db@78.47.204.74:8550",    // Community
	"enode://585913f8318ccfc6359b63f2e08e01f0097c4fb84098e27b533bd5ca6d9425700238c11599e8662b41f567155cfa14705d7f8ba68f3d3a809083da1a50baeec7@54.238.212.246:8550",  // Community
	"enode://827dddc71306d84509938106f650d7b8f27e329cd299a07c63fe5c9a0918b36240a0f14fb5634885c2e947a767c5371cfa931045dfa0db3e1a8a7057e6fe0ce8@54.238.212.246:8550",  // Community
	"enode://6f1f82c82960ed361e7542f83bf3b1a38901c61ad45aa526ae8088703edceac969403ab18cf6a9dcf9d42852e73397ddac410944ac391e8e1708baefe7c8b671@65.108.208.164:8550",  // Community
	"enode://83ec9c3d2bd349fa5ad93c88b045013edde59a717072d2f495466b97eac845fa01645e788f4aa85aeea76349cce9106dd97a85d70c7dab8bb50d131c703071a9@142.132.234.205:8550", // Community
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// INT Chain test network.
var TestnetBootnodes = []string{
	"enode://a12a394da7887a8fa99cf3347cd372b08502e3dd04f6e464967b5a3cdebf7234f2dd4e5fc9720961cf03224920b8b3c4c25ae23ba5d593345ea41b1b7e575b7d@101.32.74.50:8550", // INT Chain
	//"enode://dd7e1ead96067f73ed13c5a7a0e6d1cf7a92644a225478acda42c48750aff021263740481b5d1d72275fe19e4d61d064bbb3293adba0f0845d013f7ece63c75d@101.32.74.50:8551",   // Titans
	"enode://4dd8d8352900f44cc6bfd9d9c3f2d9c8d80d0b6a64b97019dc9e3ca905a4ca6474b84745e5001fc08c5b64910baeb233fcd54bda4f6f855438d7c762494977c0@129.226.59.148:8551", // Oceanus
	"enode://964f2c8d04bcc48e971bb24870035db8cfc7218bfdb4d97849421b47d9cc207fff1b89876b4ff137156fb52a262254a570396a4c06a756d2c68a8f637eaa5031@129.226.128.55:8551", // Iapetus
	"enode://65e60686f0c091b30a701b3c8e02db1aab75ad67e7d319c679a616cad38ce2dad5abb8829cd0d6041a9541387078eed12b88f5d0eb5168456a182ad2d0183e19@129.226.63.13:8551",  // Mnemosyne
}
