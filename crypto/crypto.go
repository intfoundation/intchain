// Copyright 2014 The go-ethereum Authors
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

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/rlp"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"hash"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
)

var (
	pubkeyVersion = byte(0x00)
	scriptVersion = byte(0x05)
	addressPrefix = "INT"
	bs58Str       = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz" // remove  0 I O l
	// 椭圆曲线的阶
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Keccak512 calculates and returns the Keccak512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Creates an ethereum address given the bytes and the nonce
//func CreateAddress(b common.Address, nonce uint64) common.Address {
//	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
//	return common.BytesToAddress(Keccak256(data)[12:])
//}

//
func CreateAddress(b common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	return common.StringToAddress(NewINTScriptAddr(data))
}

// CreateAddress2 creates an ethereum address given the address bytes, initial
// contract code and a salt.
//func CreateAddress2(b common.Address, salt [32]byte, code []byte) common.Address {
//	return common.BytesToAddress(Keccak256([]byte{0xff}, b.Bytes(), salt[:], Keccak256(code))[12:])
//}

func CreateAddress2(b common.Address, salt [32]byte, inithash []byte) common.Address {
	return common.BytesToAddress([]byte(NewINTScriptAddr(Keccak256([]byte{0xff}, b.Bytes(), salt[:], inithash))))
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// FromECDSA exports a private key into a binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

// TODO remove toecdsapub
func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	//x, y := Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errInvalidPubkey
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
	//return Marshal(S256(), pub.X, pub.Y)
}

// Marshal converts a point into the compressed
//func Marshal(curve elliptic.Curve, x, y *big.Int) []byte {
//	byteLen := (curve.Params().BitSize + 7) >> 3
//
//	ret := make([]byte, 1+byteLen)
//	//ret[0] = 4 // uncompressed point
//
//	xBytes := x.Bytes()
//	copy(ret[1+byteLen-len(xBytes):], xBytes)
//	//yBytes := y.Bytes()
//	//copy(ret[1+2*byteLen-len(yBytes):], yBytes)
//	str := []byte(y.String())
//
//	// y 为偶数，添加前缀2，y为奇数，添加前缀3
//	lastNum, _ := strconv.ParseInt(string(str[len(str)-1:][0]), 10, 10)
//	if (lastNum & 1) == 1 {
//		ret[0] = 3
//	} else {
//		ret[0] = 2
//	}
//	return ret
//}
// Unmarshal converts a point, serialized by Marshal, into an x, y pair.
// It is an error if the point is not in compressed form or is not on the curve.
// On error, x = nil.
//func Unmarshal(curve elliptic.Curve, data []byte) (x, y *big.Int) {
//	byteLen := (curve.Params().BitSize + 7) >> 3
//	if len(data) != 1+byteLen {
//		return
//	}
//	if !(data[0] == 2 || data[0] == 3) { // compressed form
//		return
//	}
//	p := curve.Params().P
//	b := curve.Params().B
//	x = new(big.Int).SetBytes(data[1 : 1+byteLen])
//	fmt.Printf("Unmarshal x=%v\n", x)
//	//y = new(big.Int).SetBytes(data[1+byteLen:])
//
//	// 由 x 计算出 y，y² mod p = x³ - 3x + b (mod p)
//	//y2 := new(big.Int).Mul(y, y)
//	//y2.Mod(y2, curve.Params().P)
//
//	x3 := new(big.Int).Mul(x, x)
//	x3.Mul(x3, x)
//
//	threeX := new(big.Int).Lsh(x, 1)
//	threeX.Add(threeX, x)
//
//	x3.Sub(x3, threeX)
//	x3.Add(x3, b)
//
//	y = big.NewInt(0)
//	//todo 计算的 y的值不对
//	y.Sqrt(x3)
//
//	//return x3.Cmp(y2) == 0
//	//y = x3.Sqrt(x3)
//	fmt.Printf("Unmarshal y=%v\n", y)
//	fmt.Printf("Unmarshal ybytes=%v\n", y.Bytes())
//
//
//	if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
//		return nil, nil
//	}
//	if !curve.IsOnCurve(x, y) {
//		return nil, nil
//	}
//	return
//}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToECDSA(b)
}

// LoadECDSA loads a secp256k1 private key from the given file.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	return ToECDSA(key)
}

// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1halfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	return r.Cmp(secp256k1N) < 0 && s.Cmp(secp256k1N) < 0 && (v == 0 || v == 1)
}

//func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
//	pubBytes := FromECDSAPub(&p)
//	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
//}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.StringToAddress(NewINTScriptAddr(pubBytes))
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

// P2PH 暂时不用，因为长度不固定
func NewINTPubkeyAddr(pubkey []byte) string {
	input := Hash160(pubkey)

	return encodeAddress(input, pubkeyVersion)
}

// P2SH
func NewINTScriptAddr(script []byte) string {
	input := Hash160(script)
	strArray := strings.Split(encodeAddress(input, scriptVersion), "")
	return addressPrefix + strings.Join(strArray[:29], "")
}

// check INT address is validate or not
func ValidateINTAddr(input string) bool {
	inputByte := []byte(input)
	if len(inputByte) != 32 {
		return false
	}

	if inputByte[0] != 'I' || inputByte[1] != 'N' || inputByte[2] != 'T' || inputByte[3] != '3' {
		return false
	}

	inputArray := strings.Split(input, "")[3:]

	for _, v := range inputArray {
		if !strings.Contains(bs58Str, v) {
			return false
		}
	}

	return true
}

// encodeAddress returns a human-readable payment address given a ripemd160 hash
// and netID which encodes the bitcoin network and address type.  It is used
// in both pay-to-pubkey-hash (P2PKH) and pay-to-script-hash (P2SH) address
// encoding.
func encodeAddress(hash160 []byte, version byte) string {
	// Format is 1 byte for a network and address class (i.e. P2PKH vs
	// P2SH), 20 bytes for a RIPEMD160 hash, and 4 bytes of checksum.
	return base58.CheckEncode(hash160, version)
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}
