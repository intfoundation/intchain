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

package intapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/intfoundation/intchain/accounts/abi"
	"github.com/intfoundation/intchain/consensus"
	"github.com/intfoundation/intchain/consensus/ipbft/epoch"
	"github.com/intfoundation/intchain/core/state"
	"math/big"
	"strings"
	"time"

	goCrypto "github.com/intfoundation/go-crypto"
	"github.com/intfoundation/intchain/accounts"
	"github.com/intfoundation/intchain/accounts/keystore"
	"github.com/intfoundation/intchain/common"
	"github.com/intfoundation/intchain/common/hexutil"
	"github.com/intfoundation/intchain/common/math"
	"github.com/intfoundation/intchain/core"
	"github.com/intfoundation/intchain/core/rawdb"
	"github.com/intfoundation/intchain/core/types"
	"github.com/intfoundation/intchain/core/vm"
	"github.com/intfoundation/intchain/crypto"
	intAbi "github.com/intfoundation/intchain/intabi/abi"
	"github.com/intfoundation/intchain/log"
	"github.com/intfoundation/intchain/p2p"
	"github.com/intfoundation/intchain/params"
	"github.com/intfoundation/intchain/rlp"
	"github.com/intfoundation/intchain/rpc"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	defaultGasPrice          = 5000 * params.GWei
	updateValidatorThreshold = 25
)

var (
	minimumRegisterAmount = math.MustParseBig256("1000000000000000000000000") //  1000000 * e18

	//maxCandidateNumber = 1000

	maxDelegationAddresses = 1000

	maxEditValidatorLength = 100
)

// PublicINTChainAPI provides an API to access intchain related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicINTChainAPI struct {
	b Backend
}

// NewPublicINTChainAPI creates a new intchain protocol API.
func NewPublicINTChainAPI(b Backend) *PublicINTChainAPI {
	return &PublicINTChainAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicINTChainAPI) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	price, err := s.b.SuggestPrice(ctx)
	return (*hexutil.Big)(price), err
}

// ProtocolVersion returns the current intchain protocol version this node supports
func (s *PublicINTChainAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicINTChainAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	if progress.CurrentBlock >= progress.HighestBlock {
		return false, nil
	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(progress.StartingBlock),
		"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock":  hexutil.Uint64(progress.HighestBlock),
		"pulledStates":  hexutil.Uint64(progress.PulledStates),
		"knownStates":   hexutil.Uint64(progress.KnownStates),
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.
func (s *PublicTxPoolAPI) Content() map[string]map[string]map[string]*RPCTransaction {
	content := map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	pending, queue := s.b.Stats()
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(pending),
		"queued":  hexutil.Uint(queue),
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.
func (s *PublicTxPoolAPI) Inspect() map[string]map[string]map[string]string {
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *types.Transaction) string {
		if to := tx.To(); to != nil {
			return fmt.Sprintf("%s: %v wei + %v gas × %v wei", tx.To().Hex(), tx.Value(), tx.Gas(), tx.GasPrice())
		}
		return fmt.Sprintf("contract creation: %v wei + %v gas × %v wei", tx.Value(), tx.Gas(), tx.GasPrice())
	}
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	b         Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.AccountManager(),
		nonceLock: nonceLock,
		b:         b,
	}
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (s *PrivateAccountAPI) ListAccounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// rawWallet is a JSON representation of an accounts.Wallet interface, with its
// data contents extracted into plain fields.
type rawWallet struct {
	URL      string             `json:"url"`
	Status   string             `json:"status"`
	Failure  string             `json:"failure,omitempty"`
	Accounts []accounts.Account `json:"accounts,omitempty"`
}

// ListWallets will return a list of wallets this node manages.
func (s *PrivateAccountAPI) ListWallets() []rawWallet {
	wallets := make([]rawWallet, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		status, failure := wallet.Status()

		raw := rawWallet{
			URL:      wallet.URL().String(),
			Status:   status,
			Accounts: wallet.Accounts(),
		}
		if failure != nil {
			raw.Failure = failure.Error()
		}
		wallets = append(wallets, raw)
	}
	return wallets
}

// OpenWallet initiates a hardware wallet opening procedure, establishing a USB
// connection and attempting to authenticate via the provided passphrase. Note,
// the method may return an extra challenge requiring a second open (e.g. the
// Trezor PIN matrix challenge).
func (s *PrivateAccountAPI) OpenWallet(url string, passphrase *string) error {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return err
	}
	pass := ""
	if passphrase != nil {
		pass = *passphrase
	}
	return wallet.Open(pass)
}

// DeriveAccount requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
func (s *PrivateAccountAPI) DeriveAccount(url string, path string, pin *bool) (accounts.Account, error) {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return accounts.Account{}, err
	}
	derivPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return accounts.Account{}, err
	}
	if pin == nil {
		pin = new(bool)
	}
	return wallet.Derive(derivPath, *pin)
}

// NewAccount will create a new account and returns the address for the new account.
func (s *PrivateAccountAPI) NewAccount(password string) (common.Address, error) {
	acc, err := fetchKeystore(s.am).NewAccount(password)
	if err == nil {
		return acc.Address, nil
	}
	return common.Address{}, err
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
func (s *PrivateAccountAPI) ImportRawKey(privkey string, password string) (common.Address, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (s *PrivateAccountAPI) UnlockAccount(addr common.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

// LockAccount will lock the account associated with the given address when it's unlocked.
func (s *PrivateAccountAPI) LockAccount(addr common.Address) bool {
	return fetchKeystore(s.am).Lock(addr) == nil
}

// signTransactions sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, if applicable,
// and release it after the transaction has been submitted to the tx pool
func (s *PrivateAccountAPI) signTransaction(ctx context.Context, args SendTxArgs, passwd string) (*types.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}
	wallet, err := s.am.Find(account)
	if err != nil {
		return nil, err
	}
	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	return wallet.SignTxWithPassphrase(account, passwd, tx, chainID)
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails.
func (s *PrivateAccountAPI) SendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
	fmt.Printf("transaction args PrivateAccountAPI args %v\n", args)
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}
	signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

// SignTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails. The transaction is returned in RLP-form, not broadcast
// to other nodes
func (s *PrivateAccountAPI) SignTransaction(ctx context.Context, args SendTxArgs, passwd string) (*SignTransactionResult, error) {
	// No need to obtain the noncelock mutex, since we won't be sending this
	// tx into the transaction pool, but right back to the user
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, signed}, nil
}

// signHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//   keccak256("\x19INT Chain Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19INT Chain Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// Sign calculates an INT Chain ECDSA signature for:
// keccack256("\x19INT Chain Signed Message:\n" + len(message) + message))
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The key used to calculate the signature is decrypted with the given password.
//
// https://github.com/intfoundation/intchain/wiki/Management-APIs#personal_sign
func (s *PrivateAccountAPI) Sign(ctx context.Context, data hexutil.Bytes, addr common.Address, passwd string) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Assemble sign the data with the wallet
	signature, err := wallet.SignHashWithPassphrase(account, passwd, signHash(data))
	if err != nil {
		return nil, err
	}
	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19INT Chain Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be 27 or 28 for legacy reasons.
//
// https://github.com/intfoundation/intchain/wiki/Management-APIs#personal_ecRecover
func (s *PrivateAccountAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid INT Chain signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(signHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

// SignAndSendTransaction was renamed to SendTransaction. This method is deprecated
// and will be removed in the future. It primary goal is to give clients time to update.
func (s *PrivateAccountAPI) SignAndSendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
	return s.SendTransaction(ctx, args, passwd)
}

// PublicBlockChainAPI provides an API to access the INT blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new INT blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// ChainId returns the chainID value for transaction replay protection.
func (s *PublicBlockChainAPI) ChainId() *hexutil.Big {
	return (*hexutil.Big)(s.b.ChainConfig().ChainId)
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) BlockNumber() hexutil.Uint64 {
	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return hexutil.Uint64(header.Number.Uint64())
}

// GetBalance returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Big, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	return (*hexutil.Big)(state.GetBalance(address)), state.Error()
}

//func (s *PublicBlockChainAPI) GetCandidateSetByBlockNumber(ctx context.Context, blockNr rpc.BlockNumber) ([]common.Address, error) {
//	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
//	if state == nil || err != nil {
//		return nil, err
//	}
//
//	var candidateList = make([]common.Address, 0)
//
//	for addr := range state.GetCandidateSet() {
//		candidateList = append(candidateList, addr)
//	}
//
//	return candidateList, nil
//}

type ProxiedDetail struct {
	ProxiedBalance        *hexutil.Big `json:"proxiedBalance"`
	DepositProxiedBalance *hexutil.Big `json:"depositProxiedBalance"`
	PendingRefundBalance  *hexutil.Big `json:"pendingRefundBalance"`
}

// GetBalanceDetail returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalanceDetail(ctx context.Context, address common.Address, blockNr rpc.BlockNumber, fullDetail bool) (map[string]interface{}, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}

	fields := map[string]interface{}{
		"balance":               (*hexutil.Big)(state.GetBalance(address)),
		"depositBalance":        (*hexutil.Big)(state.GetDepositBalance(address)),
		"delegateBalance":       (*hexutil.Big)(state.GetDelegateBalance(address)),
		"proxiedBalance":        (*hexutil.Big)(state.GetTotalProxiedBalance(address)),
		"depositProxiedBalance": (*hexutil.Big)(state.GetTotalDepositProxiedBalance(address)),
		"pendingRefundBalance":  (*hexutil.Big)(state.GetTotalPendingRefundBalance(address)),
		"rewardBalance":         (*hexutil.Big)(state.GetTotalRewardBalance(address)),
	}

	if fullDetail {
		proxiedDetail := make(map[common.Address]ProxiedDetail)
		state.ForEachProxied(address, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
			proxiedDetail[key] = ProxiedDetail{
				ProxiedBalance:        (*hexutil.Big)(proxiedBalance),
				DepositProxiedBalance: (*hexutil.Big)(depositProxiedBalance),
				PendingRefundBalance:  (*hexutil.Big)(pendingRefundBalance),
			}
			return true
		})

		fields["proxiedDetail"] = proxiedDetail

		rewardDetail := make(map[common.Address]*hexutil.Big)
		state.ForEachReward(address, func(key common.Address, rewardBalance *big.Int) bool {
			rewardDetail[key] = (*hexutil.Big)(rewardBalance)
			return true
		})

		fields["rewardDetail"] = rewardDetail
	}
	return fields, state.Error()
}

type EpochLabel uint64

func (e EpochLabel) MarshalText() ([]byte, error) {
	output := fmt.Sprintf("epoch_%d", e)
	return []byte(output), nil
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetUncleByBlockNumberAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", blockNr, "hash", block.Hash(), "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleByBlockHashAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", block.Number(), "hash", blockHash, "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleCountByBlockNumber returns number of uncles in the block for the given block number
func (s *PublicBlockChainAPI) GetUncleCountByBlockNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetUncleCountByBlockHash returns number of uncles in the block for the given block hash
func (s *PublicBlockChainAPI) GetUncleCountByBlockHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	code := state.GetCode(address)
	return code, state.Error()
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, common.HexToHash(key))
	return res[:], state.Error()
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      hexutil.Uint64  `json:"gas"`
	GasPrice hexutil.Big     `json:"gasPrice"`
	Value    hexutil.Big     `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
}

func (s *PublicBlockChainAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) (*core.ExecutionResult, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	// Set sender address or use a default if none specified
	addr := args.From
	if addr == (common.Address{}) {
		if wallets := s.b.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				addr = accounts[0].Address
			}
		}
	}
	// Set default gas & gas price if none were set
	gas, gasPrice := uint64(args.Gas), args.GasPrice.ToInt()
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}

	// Create new call message
	msg := types.NewMessage(addr, args.To, 0, args.Value.ToInt(), gas, gasPrice, args.Data, false)

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Get a new instance of the EVM.
	evm, vmError, err := s.b.GetEVM(ctx, msg, state, header, vmCfg)
	if err != nil {
		return nil, err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	result, _, err := core.ApplyMessageEx(evm, msg, gp)
	if err := vmError(); err != nil {
		return nil, err
	}
	// If the timer caused an abort, return an appropriate error message
	//if evm.Cancelled() {
	//	return nil, fmt.Errorf("execution aborted (timeout = %v)", timeout)
	//}
	if err != nil {
		return result, fmt.Errorf("err: %w (supplied gas %d)", err, msg.Gas())
	}

	return result, err
}

func newRevertError(result *core.ExecutionResult) *revertError {
	reason, errUnpack := abi.UnpackRevert(result.Revert())
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &revertError{
		error:  err,
		reason: hexutil.Encode(result.Revert()),
	}
}

// revertError is an API error that encompassas an EVM revertal with JSON error
// code and a binary data blob.
type revertError struct {
	error
	reason string // revert reason hex encoded
}

// ErrorCode returns the JSON error code for a revertal.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *revertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *revertError) ErrorData() interface{} {
	return e.reason
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	result, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	//return (hexutil.Bytes)(result), err

	if err != nil {
		return nil, err
	}
	// If the result contains a revert reason, try to unpack and return it.
	if len(result.Revert()) > 0 {
		return nil, newRevertError(result)
	}
	return result.Return(), result.Err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)

	functionType, e := intAbi.FunctionTypeFromId(args.Data[:4])
	if e == nil && functionType != intAbi.Unknown {
		fmt.Printf("intchain inner contract tx, address: %v, functionType: %v\n", args.To.Hex(), functionType)
		return hexutil.Uint64(functionType.RequiredGas()), nil
	}

	if uint64(args.Gas) >= params.TxGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block, err := s.b.BlockByNumber(ctx, rpc.PendingBlockNumber)
		if err != nil {
			return 0, err
		}
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *core.ExecutionResult, error) {
		args.Gas = hexutil.Uint64(gas)

		result, err := s.doCall(ctx, args, rpc.PendingBlockNumber, vm.Config{}, 0)
		//if err != nil || failed {
		//	return false
		//}
		//return true
		if err != nil {
			if errors.Is(err, core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, err // Bail out
		}
		return result.Failed(), result, nil
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		//mid := (hi + lo) / 2
		//if !executable(mid) {
		//	lo = mid
		//} else {
		//	hi = mid
		//}
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)

		// If the error is not nil(consensus error), it means the provided message
		// call or transaction will never be accepted no matter how much gas it is
		// assigned. Return the error directly, don't struggle any more.
		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		//if !executable(hi) {
		//	return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		//}
		failed, result, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if result != nil && result.Err != vm.ErrOutOfGas {
				if len(result.Revert()) > 0 {
					return 0, newRevertError(result)
				}
				return 0, result.Err
			}
			// Otherwise, the specified gas cap is too low
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", cap)
		}
	}
	return hexutil.Uint64(hi), nil
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   error              `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// formatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.Err,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", math.PaddedBigBytes(stackValue, 32))
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

// rpcOutputBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number": (*hexutil.Big)(head.Number),
		//"mainchainNumber":  (*hexutil.Big)(head.MainChainNumber),
		"hash":             b.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"totalDifficulty":  (*hexutil.Big)(s.b.GetTd(b.Hash())),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(b.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        (*hexutil.Big)(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}

	if inclTx {
		formatTx := func(tx *types.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}

		if fullTx {
			formatTx = func(tx *types.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}

		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range b.Transactions() {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	uncles := b.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *types.Transaction) *RPCTransaction {
	return newRPCTransaction(tx, common.Hash{}, 0, 0)
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *types.Block, index uint64) hexutil.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	blob, _ := rlp.EncodeToBytes(txs[index])
	return blob
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	nonce := state.GetNonce(address)
	return (*hexutil.Uint64)(&nonce), state.Error()
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicTransactionPoolAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) *RPCTransaction {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return newRPCTransaction(tx, blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return newRPCPendingTransaction(tx)
	}
	// Transaction unknown, return as such
	return nil
}

// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	var tx *types.Transaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = rawdb.ReadTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to RLP and return
	return rlp.EncodeToBytes(tx)
}

type Log struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address string `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data string `json:"data" gencodec:"required"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockNumber uint64 `json:"blockNumber"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required"`
	// index of the transaction in the block
	TxIndex uint `json:"transactionIndex" gencodec:"required"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash"`
	// index of the log in the receipt
	Index uint `json:"logIndex" gencodec:"required"`

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed"`
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]

	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(blockNumber),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(index),
		"from":              from,
		"to":                tx.To(),
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
	}

	// Assign receipt status or post state.
	if len(receipt.PostState) > 0 {
		fields["root"] = hexutil.Bytes(receipt.PostState)
	} else {
		fields["status"] = hexutil.Uint(receipt.Status)
	}
	if receipt.Logs == nil {
		fields["logs"] = [][]*types.Log{}
	}

	// If the ContractAddress is 32 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = receipt.ContractAddress
	}
	return fields, nil
}

// sign is a helper function that signs a transaction with the private key of the given address.
func (s *PublicTransactionPoolAPI) sign(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Request the wallet to sign the transaction
	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	return wallet.SignTxWithAddress(account, tx, chainID)
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {

	var function = intAbi.Unknown
	if intAbi.IsIntChainContractAddr(args.To) {
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`intchain contract without any data provided`)
		}

		var err error
		function, err = intAbi.FunctionTypeFromId(input[:4])
		if err != nil {
			return err
		}
	}

	// force GasLimit to 0 for DepositInChildChain/WithdrawFromMainChain/SaveDataToMainChain in order to avoid being dropped by TxPool.
	if function == intAbi.DepositInChildChain || function == intAbi.WithdrawFromMainChain || function == intAbi.SaveDataToMainChain {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 0
	} else {
		if args.Gas == nil {
			args.Gas = new(hexutil.Uint64)
			*(*uint64)(args.Gas) = 90000
		}
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`Both "data" and "input" are set and not equal. Please use "input" to pass transaction call data.`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}

	if args.Nonce == nil {
		nonce, err := b.GetPoolNonce(ctx, args.From)
		if err != nil {
			return err
		}
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}

	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *types.Transaction) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	if tx.To() == nil {
		signer := types.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number())
		from, err := types.Sender(signer, tx)
		if err != nil {
			return common.Hash{}, err
		}
		addr := crypto.CreateAddress(from, tx.Nonce())
		log.Info("Submitted contract creation", "fullhash", tx.Hash().Hex(), "contract", addr.Hex())
	} else {
		log.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient", tx.To())
	}
	return tx.Hash(), nil
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	fmt.Printf("transaction args PublicTransactionPoolAPI args %v\n", args)
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	signed, err := wallet.SignTxWithAddress(account, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, signed)
}

func SendTransaction(ctx context.Context, args SendTxArgs, am *accounts.Manager, b Backend, nonceLock *AddrLocker) (common.Hash, error) {
	fmt.Printf("transaction args PublicTransactionPoolAPI args %v\n", args)
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := am.Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		nonceLock.LockAddr(args.From)
		defer nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := b.ChainConfig(); config.IsEIP155(b.CurrentBlock().Number()) {
		chainID = config.ChainId
	}
	signed, err := wallet.SignTxWithAddress(account, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, b, signed)
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx)
}

// Sign calculates an ECDSA signature for:
// keccack256("\x19INT Chain Signed Message:\n" + len(message) + message).
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The account associated with addr must be unlocked.
//
func (s *PublicTransactionPoolAPI) Sign(addr common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the requested hash with the wallet
	signature, err := wallet.SignHash(account, signHash(data))
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signature, err
}

// SignTransactionResult represents a RLP encoded signed transaction.
type SignTransactionResult struct {
	Raw hexutil.Bytes      `json:"raw"`
	Tx  *types.Transaction `json:"tx"`
}

// SignTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) SignTransaction(ctx context.Context, args SendTxArgs) (*SignTransactionResult, error) {
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	tx, err := s.sign(args.From, args.toTransaction())
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, tx}, nil
}

// PendingTransactions returns the transactions that are in the transaction pool and have a from address that is one of
// the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}

	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(tx.ChainId())
		}
		from, _ := types.Sender(signer, tx)
		if _, err := s.b.AccountManager().Find(accounts.Account{Address: from}); err == nil {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}

// Resend accepts an existing transaction and a new gas price and limit. It will remove
// the given transaction from the pool and reinsert it with the new gas price and limit.
func (s *PublicTransactionPoolAPI) Resend(ctx context.Context, sendArgs SendTxArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error) {
	if sendArgs.Nonce == nil {
		return common.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}
	if err := sendArgs.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	matchTx := sendArgs.toTransaction()
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return common.Hash{}, err
	}

	for _, p := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if p.Protected() {
			signer = types.NewEIP155Signer(p.ChainId())
		}
		wantSigHash := signer.Hash(matchTx)

		if pFrom, err := types.Sender(signer, p); err == nil && pFrom == sendArgs.From && signer.Hash(p) == wantSigHash {
			// Match. Re-sign and send the transaction.
			if gasPrice != nil && (*big.Int)(gasPrice).Sign() != 0 {
				sendArgs.GasPrice = gasPrice
			}
			if gasLimit != nil && *gasLimit != 0 {
				sendArgs.Gas = gasLimit
			}
			signedTx, err := s.sign(sendArgs.From, sendArgs.toTransaction())
			if err != nil {
				return common.Hash{}, err
			}
			if err = s.b.SendTx(ctx, signedTx); err != nil {
				return common.Hash{}, err
			}
			return signedTx.Hash(), nil
		}
	}

	return common.Hash{}, fmt.Errorf("Transaction %#x not found", matchTx.Hash())
}

// PublicDebugAPI is the collection of INT Chain APIs exposed over the public
// debugging endpoint.
type PublicDebugAPI struct {
	b Backend
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the INT Chain service.
func NewPublicDebugAPI(b Backend) *PublicDebugAPI {
	return &PublicDebugAPI{b: b}
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (api *PublicDebugAPI) GetBlockRlp(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	encoded, err := rlp.EncodeToBytes(block)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", encoded), nil
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (api *PublicDebugAPI) PrintBlock(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return block.String(), nil
}

// PrivateDebugAPI is the collection of INT Chain APIs exposed over the private
// debugging endpoint.
type PrivateDebugAPI struct {
	b Backend
}

// NewPrivateDebugAPI creates a new API definition for the private debug methods
// of the INT Chain service.
func NewPrivateDebugAPI(b Backend) *PrivateDebugAPI {
	return &PrivateDebugAPI{b: b}
}

// ChaindbProperty returns leveldb properties of the chain database.
func (api *PrivateDebugAPI) ChaindbProperty(property string) (string, error) {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return "", fmt.Errorf("chaindbProperty does not work for memory databases")
	}
	if property == "" {
		property = "leveldb.stats"
	} else if !strings.HasPrefix(property, "leveldb.") {
		property = "leveldb." + property
	}
	return ldb.LDB().GetProperty(property)
}

func (api *PrivateDebugAPI) ChaindbCompact() error {
	for b := byte(0); b < 255; b++ {
		log.Info("Compacting chain database", "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
		if err := api.b.ChainDb().Compact([]byte{b}, []byte{b + 1}); err != nil {
			log.Error("Database compaction failed", "err", err)
			return err
		}
	}
	return nil
}

// SetHead rewinds the head of the blockchain to a previous block.
func (api *PrivateDebugAPI) SetHead(number hexutil.Uint64) {
	api.b.SetHead(uint64(number))
}

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	net            *p2p.Server
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net *p2p.Server, networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{net, networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (s *PublicNetAPI) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *PublicNetAPI) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.PeerCount())
}

// Version returns the current intchain protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

type PublicINTAPI struct {
	am        *accounts.Manager
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicINTAPI creates a new INT API instance.
func NewPublicINTAPI(b Backend, nonceLock *AddrLocker) *PublicINTAPI {
	return &PublicINTAPI{b.AccountManager(), b, nonceLock}
}

func (s *PublicINTAPI) SignAddress(from common.Address, consensusPrivateKey hexutil.Bytes) (goCrypto.Signature, error) {
	if len(consensusPrivateKey) != 32 {
		return nil, errors.New("invalid consensus private key")
	}

	var blsPriv goCrypto.BLSPrivKey
	copy(blsPriv[:], consensusPrivateKey)

	blsSign := blsPriv.Sign(from.Bytes())

	return blsSign, nil
}

func (api *PublicINTAPI) WithdrawReward(ctx context.Context, from common.Address, delegateAddress common.Address, amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.WithdrawReward.String(), delegateAddress, (*big.Int)(amount))
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.WithdrawReward.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) Delegate(ctx context.Context, from, candidate common.Address, amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.Delegate.String(), candidate)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.Delegate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    amount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) UnDelegate(ctx context.Context, from, candidate common.Address, amount *hexutil.Big, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.UnDelegate.String(), candidate, (*big.Int)(amount))
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.UnDelegate.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) Register(ctx context.Context, from common.Address, registerAmount *hexutil.Big, pubkey goCrypto.BLSPubKey, signature hexutil.Bytes, commission uint8, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.Register.String(), pubkey.Bytes(), signature, commission)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.Register.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    registerAmount,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) UnRegister(ctx context.Context, from common.Address, gasPrice *hexutil.Big) (common.Hash, error) {

	input, err := intAbi.ChainABI.Pack(intAbi.UnRegister.String())
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.UnRegister.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}
	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) CheckCandidate(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (map[string]interface{}, error) {
	state, _, err := api.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}

	fields := map[string]interface{}{
		"candidate":  state.IsCandidate(address),
		"commission": state.GetCommission(address),
	}
	return fields, state.Error()
}

func (api *PublicINTAPI) SetCommission(ctx context.Context, from common.Address, commission uint8, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.SetCommission.String(), commission)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.SetCommission.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) EditValidator(ctx context.Context, from common.Address, moniker, website string, identity string, details string, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.EditValidator.String(), moniker, website, identity, details)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.EditValidator.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func (api *PublicINTAPI) SetAddress(ctx context.Context, from, fAddress common.Address, gasPrice *hexutil.Big) (common.Hash, error) {
	input, err := intAbi.ChainABI.Pack(intAbi.SetAddress.String(), fAddress)
	if err != nil {
		return common.Hash{}, err
	}

	defaultGas := intAbi.SetAddress.RequiredGas()

	args := SendTxArgs{
		From:     from,
		To:       &intAbi.ChainContractMagicAddr,
		Gas:      (*hexutil.Uint64)(&defaultGas),
		GasPrice: gasPrice,
		Value:    nil,
		Input:    (*hexutil.Bytes)(&input),
		Nonce:    nil,
	}

	return SendTransaction(ctx, args, api.am, api.b, api.nonceLock)
}

func init() {
	// Withdraw reward
	core.RegisterValidateCb(intAbi.WithdrawReward, withdrawRewardValidateCb)
	core.RegisterApplyCb(intAbi.WithdrawReward, withdrawRewardApplyCb)

	// Delegate
	core.RegisterValidateCb(intAbi.Delegate, delegateValidateCb)
	core.RegisterApplyCb(intAbi.Delegate, delegateApplyCb)

	// Cancel Delegate
	core.RegisterValidateCb(intAbi.UnDelegate, unDelegateValidateCb)
	core.RegisterApplyCb(intAbi.UnDelegate, unDelegateApplyCb)

	// Register
	core.RegisterValidateCb(intAbi.Register, registerValidateCb)
	core.RegisterApplyCb(intAbi.Register, registerApplyCb)

	// Cancel Register
	core.RegisterValidateCb(intAbi.UnRegister, unRegisterValidateCb)
	core.RegisterApplyCb(intAbi.UnRegister, unRegisterApplyCb)

	// Set Commission
	core.RegisterValidateCb(intAbi.SetCommission, setCommisstionValidateCb)
	core.RegisterApplyCb(intAbi.SetCommission, setCommisstionApplyCb)

	// Edit Validator
	core.RegisterValidateCb(intAbi.EditValidator, editValidatorValidateCb)

	// UnForbidden
	//core.RegisterValidateCb(intAbi.UnForbidden, unForbiddenValidateCb)
	//core.RegisterApplyCb(intAbi.UnForbidden, unForbiddenApplyCb)

	// Set Address
	core.RegisterValidateCb(intAbi.SetAddress, setAddressValidateCb)
	core.RegisterApplyCb(intAbi.SetAddress, setAddressApplyCb)
}

func withdrawRewardValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, err := withDrawRewardValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func withdrawRewardApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	from := derivedAddressFromTx(tx)

	args, err := withDrawRewardValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	//reward := state.GetRewardBalanceByDelegateAddress(from, args.DelegateAddress)
	state.SubRewardBalanceByDelegateAddress(from, args.DelegateAddress, args.Amount)
	state.AddBalance(from, args.Amount)

	return nil
}

func withDrawRewardValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.WithdrawRewardArgs, error) {

	var args intAbi.WithdrawRewardArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.WithdrawReward.String(), data[4:]); err != nil {
		return nil, err
	}

	reward := state.GetRewardBalanceByDelegateAddress(from, args.DelegateAddress)

	if reward.Sign() < 1 {
		return nil, fmt.Errorf("have no reward to withdraw")
	}

	if args.Amount.Sign() == -1 {
		return nil, fmt.Errorf("widthdraw amount can not be negative")
	}

	if args.Amount.Cmp(reward) == 1 {
		return nil, fmt.Errorf("reward balance not enough, withdraw amount %v, but balance %v, delegate address %v", args.Amount, reward, args.DelegateAddress)
	}
	return &args, nil
}

// register and unregister
func registerValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := registerValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func registerApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := registerValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// block height validation
	verror = updateValidation(bc)
	if verror != nil {
		return verror
	}

	amount := tx.Value()
	// Add minimum register amount to self
	state.SubBalance(from, amount)
	state.AddDelegateBalance(from, amount)
	state.AddProxiedBalanceByUser(from, from, amount)
	// Become a Candidate

	var blsPK goCrypto.BLSPubKey
	copy(blsPK[:], args.Pubkey)
	if verror != nil {
		return verror
	}
	state.ApplyForCandidate(from, blsPK.KeyString(), args.Commission)

	// mark address candidate
	//state.MarkAddressCandidate(from)

	verror = updateNextEpochValidatorVoteSet(tx, state, bc, from, ops)
	if verror != nil {
		return verror
	}

	return nil
}

func registerValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.RegisterArgs, error) {
	//candidateSet := state.GetCandidateSet()
	//if len(candidateSet) > maxCandidateNumber {
	//	return nil, core.ErrMaxCandidate
	//}

	// Check cleaned Candidate
	if !state.IsCleanAddress(from) {
		return nil, core.ErrAlreadyCandidate
	}

	// Check minimum register amount
	if tx.Value().Cmp(minimumRegisterAmount) == -1 {
		return nil, core.ErrMinimumRegisterAmount
	}

	var args intAbi.RegisterArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.Register.String(), data[4:]); err != nil {
		return nil, err
	}

	if err := goCrypto.CheckConsensusPubKey(from, args.Pubkey, args.Signature); err != nil {
		return nil, err
	}

	// Check Commission Range
	if args.Commission > 100 {
		return nil, core.ErrCommission
	}

	// Annual/SemiAnnual supernode can not become candidate
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(from.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return nil, core.ErrCannotCandidate
	}

	return &args, nil
}

func unRegisterValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	verror := unRegisterValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func unRegisterApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	verror := unRegisterValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// Do job
	allRefund := true
	// Refund all the amount back to users
	state.ForEachProxied(from, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
		// Refund Proxied Amount
		state.SubProxiedBalanceByUser(from, key, proxiedBalance)
		state.SubDelegateBalance(key, proxiedBalance)
		state.AddBalance(key, proxiedBalance)

		// Refund Deposit to PendingRefund if deposit > 0
		if depositProxiedBalance.Sign() > 0 {
			allRefund = false
			//Calculate the refunding amount user canceled by oneself before
			refunded := state.GetPendingRefundBalanceByUser(from, key)
			//Add the rest to refunding balance
			state.AddPendingRefundBalanceByUser(from, key, new(big.Int).Sub(depositProxiedBalance, refunded))
			// TODO Add Pending Refund Set, Commit the Refund Set
			state.MarkDelegateAddressRefund(from)
		}
		return true
	})

	state.CancelCandidate(from, allRefund)

	return nil
}

func unRegisterValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	// Check already Candidate
	if !state.IsCandidate(from) {
		return core.ErrNotCandidate
	}

	// Forbidden candidate can't unregister
	//if state.GetForbidden(from) {
	//	return core.ErrForbiddenUnRegister
	//}

	// Super node can't unregister
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(from.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return core.ErrCannotUnRegister
	}

	// Check Epoch Height
	if _, err := getEpoch(bc); err != nil {
		return err
	}

	return nil
}

// delegate and unDelegate
func delegateValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := delegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func delegateApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := delegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// block height validation
	verror = updateValidation(bc)
	if verror != nil {
		return verror
	}

	// Do job
	amount := tx.Value()
	// Move Balance to delegate balance
	state.SubBalance(from, amount)
	state.AddDelegateBalance(from, amount)
	// Add Balance to Candidate's Proxied Balance
	state.AddProxiedBalanceByUser(args.Candidate, from, amount)

	verror = updateNextEpochValidatorVoteSet(tx, state, bc, args.Candidate, ops)
	if verror != nil {
		return verror
	}

	return nil
}

func delegateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.DelegateArgs, error) {
	// Check minimum delegate amount
	if tx.Value().Sign() == -1 {
		return nil, core.ErrDelegateAmount
	}

	var args intAbi.DelegateArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.Delegate.String(), data[4:]); err != nil {
		return nil, err
	}

	// Check Candidate
	if !state.IsCandidate(args.Candidate) {
		return nil, core.ErrNotCandidate
	}

	depositBalance := state.GetDepositProxiedBalanceByUser(args.Candidate, from)
	if depositBalance.Sign() == 0 {
		// Check if exceed the limit of delegated addresses
		// if exceed the limit of delegation address number, return error
		delegatedAddressNumber := state.GetProxiedAddressNumber(args.Candidate)
		if delegatedAddressNumber >= maxDelegationAddresses {
			return nil, core.ErrExceedDelegationAddressLimit
		}
	}

	// If Candidate is supernode, only allow to increase the stack(whitelist proxied list), not allow to create the new stack
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(args.Candidate.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		if depositBalance.Sign() == 0 {
			return nil, core.ErrCannotDelegate
		}
	}

	// Check Epoch Height
	//if _, err := getEpoch(bc); err != nil {
	//	return nil, err
	//}
	return &args, nil
}

func unDelegateValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, verror := unDelegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}
	return nil
}

func unDelegateApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	// Validate first
	from := derivedAddressFromTx(tx)
	args, verror := unDelegateValidation(from, tx, state, bc)
	if verror != nil {
		return verror
	}

	// block height validation
	verror = updateValidation(bc)
	if verror != nil {
		return verror
	}

	// Apply Logic
	// if request amount < proxied amount, refund it immediately
	// otherwise, refund the proxied amount, and put the rest to pending refund balance
	proxiedBalance := state.GetProxiedBalanceByUser(args.Candidate, from)
	var immediatelyRefund *big.Int
	if args.Amount.Cmp(proxiedBalance) <= 0 {
		immediatelyRefund = args.Amount
	} else {
		immediatelyRefund = proxiedBalance
		restRefund := new(big.Int).Sub(args.Amount, proxiedBalance)
		state.AddPendingRefundBalanceByUser(args.Candidate, from, restRefund)
		// TODO Add Pending Refund Set, Commit the Refund Set
		state.MarkDelegateAddressRefund(args.Candidate)
	}

	state.SubProxiedBalanceByUser(args.Candidate, from, immediatelyRefund)
	state.SubDelegateBalance(from, immediatelyRefund)
	state.AddBalance(from, immediatelyRefund)

	verror = updateNextEpochValidatorVoteSet(tx, state, bc, args.Candidate, ops)
	if verror != nil {
		return verror
	}

	return nil
}

func unDelegateValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.UnDelegateArgs, error) {

	var args intAbi.UnDelegateArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.UnDelegate.String(), data[4:]); err != nil {
		return nil, err
	}

	if args.Amount.Sign() == -1 {
		return nil, fmt.Errorf("undelegate amount can not be negative")
	}

	// Check Self Address
	if from == args.Candidate {
		return nil, core.ErrCancelSelfDelegate
	}

	// Super node Candidate can't decrease balance
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}
	if _, supernode := ep.Validators.GetByAddress(args.Candidate.Bytes()); supernode != nil && supernode.RemainingEpoch > 0 {
		return nil, core.ErrCannotUnBond
	}

	// Check Proxied Amount in Candidate Balance
	proxiedBalance := state.GetProxiedBalanceByUser(args.Candidate, from)
	depositProxiedBalance := state.GetDepositProxiedBalanceByUser(args.Candidate, from)
	pendingRefundBalance := state.GetPendingRefundBalanceByUser(args.Candidate, from)
	// net = deposit - pending refund
	netDeposit := new(big.Int).Sub(depositProxiedBalance, pendingRefundBalance)
	// available = proxied + net
	availableRefundBalance := new(big.Int).Add(proxiedBalance, netDeposit)
	if args.Amount.Cmp(availableRefundBalance) == 1 {
		return nil, core.ErrInsufficientProxiedBalance
	}

	// if left, the left must be greater than the min delegate amount
	//remainingBalance := new(big.Int).Sub(availableRefundBalance, args.Amount)
	//if remainingBalance.Sign() == 1 && remainingBalance.Cmp(minimumDelegationAmount) == -1 {
	//	return nil, core.ErrDelegateAmount
	//}

	// Check Epoch Height
	if _, err := getEpoch(bc); err != nil {
		return nil, err
	}

	return &args, nil
}

// set commission
func setCommisstionValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, err := setCommissionValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func setCommisstionApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	from := derivedAddressFromTx(tx)
	args, err := setCommissionValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	state.SetCommission(from, args.Commission)

	return nil
}

func setCommissionValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.SetCommissionArgs, error) {
	if !state.IsCandidate(from) {
		return nil, core.ErrNotCandidate
	}

	var args intAbi.SetCommissionArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.SetCommission.String(), data[4:]); err != nil {
		return nil, err
	}

	if args.Commission > 100 {
		return nil, core.ErrCommission
	}

	return &args, nil
}

func setAddressValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	_, err := setAddressValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	return nil
}

func setAddressApplyCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, ops *types.PendingOps) error {
	from := derivedAddressFromTx(tx)
	args, err := setAddressValidation(from, tx, state, bc)
	if err != nil {
		return err
	}

	state.SetAddress(from, args.FAddress)

	return nil
}

func setAddressValidation(from common.Address, tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) (*intAbi.SetAddressArgs, error) {
	var args intAbi.SetAddressArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.SetAddress.String(), data[4:]); err != nil {
		return nil, err
	}

	return &args, nil
}

func editValidatorValidateCb(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain) error {
	from := derivedAddressFromTx(tx)
	if !state.IsCandidate(from) {
		return errors.New("you are not a validator or candidate")
	}

	var args intAbi.EditValidatorArgs
	data := tx.Data()
	if err := intAbi.ChainABI.UnpackMethodInputs(&args, intAbi.EditValidator.String(), data[4:]); err != nil {
		return err
	}

	if len([]byte(args.Details)) > maxEditValidatorLength ||
		len([]byte(args.Identity)) > maxEditValidatorLength ||
		len([]byte(args.Moniker)) > maxEditValidatorLength ||
		len([]byte(args.Website)) > maxEditValidatorLength {
		//fmt.Printf("args details length %v, identity length %v, moniker lenth %v, website length %v\n", len([]byte(args.Details)),len([]byte(args.Identity)),len([]byte(args.Moniker)),len([]byte(args.Website)))
		return fmt.Errorf("args length too long, more than %v", maxEditValidatorLength)
	}

	return nil
}

func concatCopyPreAllocate(slices [][]byte) []byte {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]byte, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}

func getEpoch(bc *core.BlockChain) (*epoch.Epoch, error) {
	var ep *epoch.Epoch
	if tdm, ok := bc.Engine().(consensus.IPBFT); ok {
		ep = tdm.GetEpoch().GetEpochByBlockNumber(bc.CurrentBlock().NumberU64())
	}

	if ep == nil {
		return nil, errors.New("epoch is nil, are you running on IPBFT Consensus Engine")
	}

	return ep, nil
}

func derivedAddressFromTx(tx *types.Transaction) (from common.Address) {
	signer := types.NewEIP155Signer(tx.ChainId())
	from, _ = types.Sender(signer, tx)
	return
}

func updateValidation(bc *core.BlockChain) error {
	ep, err := getEpoch(bc)
	if err != nil {
		return err
	}

	currHeight := bc.CurrentBlock().NumberU64()

	if currHeight <= ep.StartBlock+2 || currHeight == ep.EndBlock {
		return errors.New("incorrect block height, please retry later")
	}

	return nil
}

func updateNextEpochValidatorVoteSet(tx *types.Transaction, state *state.StateDB, bc *core.BlockChain, candidate common.Address, ops *types.PendingOps) error {
	var update bool
	ep, err := getEpoch(bc)
	if err != nil {
		return err
	}

	// calculate the net proxied balance of this candidate
	proxiedBalance := state.GetTotalProxiedBalance(candidate)
	depositProxiedBalance := state.GetTotalDepositProxiedBalance(candidate)
	pendingRefundBalance := state.GetTotalPendingRefundBalance(candidate)
	netProxied := new(big.Int).Sub(new(big.Int).Add(proxiedBalance, depositProxiedBalance), pendingRefundBalance)

	if netProxied.Sign() == -1 {
		return errors.New("validator voting power can not be negative")
	}

	//fmt.Printf("update next epoch voteset %v\n", ep.GetEpochValidatorVoteSet())
	currentEpochVoteSet := ep.GetEpochValidatorVoteSet()
	//fmt.Printf("update next epoch current epoch voteset %v\n", ep.GetEpochValidatorVoteSet())

	// whether update next epoch vote set
	if currentEpochVoteSet == nil {
		update = true
	} else {
		// if current validator size bigger than updateValidatorThreshold and the netProxied is bigger then one of the current validator voting power
		if len(currentEpochVoteSet.Votes) >= updateValidatorThreshold {
			for _, val := range currentEpochVoteSet.Votes {
				// TODO whether need compare
				if val.Amount.Cmp(netProxied) == -1 {
					update = true
					break
				}
			}
		} else {
			update = true
		}
	}

	// update is true and the address is candidate, then update next epoch validator vote set
	if update && state.IsCandidate(candidate) {
		// no need move
		// Move delegate amount first if Candidate
		//state.ForEachProxied(candidate, func(key common.Address, proxiedBalance, depositProxiedBalance, pendingRefundBalance *big.Int) bool {
		//	// Move Proxied Amount to Deposit Proxied Amount
		//	state.SubProxiedBalanceByUser(candidate, key, proxiedBalance)
		//	state.AddDepositProxiedBalanceByUser(candidate, key, proxiedBalance)
		//	return true
		//})

		var pubkey string
		pubkey = state.GetPubkey(candidate)
		pubkeyBytes := common.FromHex(pubkey)
		if pubkey == "" || len(pubkeyBytes) != 128 {
			return errors.New("wrong format of required field 'pub_key'")
		}
		var blsPK goCrypto.BLSPubKey
		copy(blsPK[:], pubkeyBytes)

		op := types.UpdateNextEpochOp{
			From:   candidate,
			PubKey: blsPK,
			Amount: netProxied,
			Salt:   "intchain",
			TxHash: tx.Hash(),
		}

		if ok := ops.Append(&op); !ok {
			return fmt.Errorf("pending ops conflict: %v", op)
		}
	}

	return nil
}
