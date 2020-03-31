// Copyright 2016 The go-ethereum Authors
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

package keystore

import (
	"fmt"
	"github.com/intfoundation/intchain/crypto"
	"io/ioutil"
	"testing"

	"github.com/intfoundation/intchain/common"
)

const (
	veryLightScryptN = 2
	veryLightScryptP = 1
)

// Tests that a json key file can be decrypted and encrypted in multiple rounds.
func TestKeyEncryptDecrypt(t *testing.T) {
	keyjson, err := ioutil.ReadFile("testdata/very-light-scrypt.json")
	if err != nil {
		t.Fatal(err)
	}
	password := ""
	address := common.HexToAddress("45dea0fb0bba44f4fcf290bba71fd57d7117cbb8")

	// Do a few rounds of decryption and encryption
	for i := 0; i < 3; i++ {
		// Try a bad password first
		if _, err := DecryptKey(keyjson, password+"bad"); err == nil {
			t.Errorf("test %d: json key decrypted with bad password", i)
		}
		// Decrypt with the correct password
		key, err := DecryptKey(keyjson, password)
		if err != nil {
			t.Fatalf("test %d: json key failed to decrypt: %v", i, err)
		}
		if key.Address != address {
			t.Errorf("test %d: key address mismatch: have %x, want %x", i, key.Address, address)
		}
		// Recrypt with a new password and start over
		password += "new data appended"
		if keyjson, err = EncryptKey(key, password, veryLightScryptN, veryLightScryptP); err != nil {
			t.Errorf("test %d: failed to recrypt key %v", i, err)
		}
	}
}

func TestEncryptKey(t *testing.T) {
	//privateKeyHex := "182e4cc598610e2e8fe3c23a9b3145c8500cb8bc1ec44d0faf4e7b3452ac5ca0"
	//key, err := crypto.HexToECDSA(privateKeyHex)
	//if err != nil {
	//	t.Fatalf("failed to decode privatekey")
	//}
	//password := "intchain"
	//keyjson, err := EncryptKey(key, password, veryLightScryptN, veryLightScryptP)
	//if err != nil {
	//	t.Errorf("failed to encrypt key %v", err)
	//}

}

func TestDecryptKey(t *testing.T) {
	keyjson, err := ioutil.ReadFile("testdata/keystore/UTC--2019-10-17T06-41-37.816846000Z--3K7YBykphE6N8jFGVbNAWfvor94i9nigU8")
	if err != nil {
		t.Fatal(err)
	}

	password := "intchain"
	address := common.StringToAddress("3K7YBykphE6N8jFGVbNAWfvor94i9nigU8")
	key, err := DecryptKey(keyjson, password)
	fmt.Printf("private key %x\n", crypto.FromECDSA(key.PrivateKey))
	if err != nil {
		t.Fatalf("json key failed to decrypt err=%v\n", err)
	}

	fmt.Printf("address are key.Address %x, address %x\n", key.Address, address)
	if key.Address != address {
		t.Errorf("address mismatch: have %v, want %v\n", key.Address.String(), address.String())
	}
}
