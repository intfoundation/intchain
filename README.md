# INT Chain

![intchain](https://raw.githubusercontent.com/intfoundation/intchain/master/docs/intchain.jpg)

INT Chain is the world's first bottom up new-generation blockchain of things (BoT) communication standard and base application platform. The ecosystem is specifically designed for easy integration with any IoT protocol.

To both improve and encourage device interconnectivity, we have built an economy driven ecosystem by providing token-incentives through a decentralized TCP/IP based architecture of IoT. This new business model, molded by IoT devices, will support an entirely new ecosystem of the Internet of Things.


## Install

### Latest Version

The latest INT Chain version for Testnet is [v4.0.01](https://github.com/intfoundation/intchain/releases/latest)

### Install `Go`


**Go 1.12.5+** is required for building and installing the INT Chain software.


Install `Go` by following the [official docs](https://golang.org/doc/install).

Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export GOBIN=$GOPATH/bin" >> ~/.bashrc
echo "export PATH=$PATH:$GOBIN" >> ~/.bashrc
source ~/.bashrc
```

Verify that `Go` has been installed successfully

```bash
go version
```

### Install `C compiler`

You can install C compiler by your favourite package manager.


### Install `intchain`

After setting up `Go` and `C compiler` correctly, you should be able to compile and run `intchain`.

Make sure that your server can access to google.com because our project depends on some libraries provided by google. (If you are not able to access google.com, you can also try to add a proxy: `export GOPROXY=https://goproxy.io`)

```bash
git clone --branch v4.0.01 https://github.com/intfoundation/intchain
cd intchain
make intchain
```

If your environment variables have set up correctly, you should not get any errors by running the above commands.
Now check your `intchain` version.

```bash
./bin/intchain version
```

## Running `intchain`

`intchain` is the entry point into the INT Chain network(main, test or private network). It can be used by other processes as a gateway into the INT Chain network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports.

We've enumerated a few common command to let you run your own `intchain` instance quickly. If you want to look over all command line options, please use `intchain --help`


### Main network

The most common condition is that users want to simply interact with the INT Chain network: create accounts; transfer funds; deploy and interact with contracts.

```bash
./bin/intchain 
```

Then you can attach to an already running `intchain` instance with `intchain attach`, and you can invoke all official methods.

```bash
./bin/intchain attach <datadir>/intchain/intchain.ipc
```



### Test network

For developers, you would like to deploy contracts for testing, but you do not want to do that with real money spending.
In other words, instead of attaching to the main network, you want to join the test network with your node, which is fully equivalent to the main network.

```bash
./bin/intchain --testnet
```

Specifying the `--testnet` flag, will reconfigure `intchain` instance a bit:
   
   * Instead of using the default data directory(`~/.intchain/intchain` on Linux for example), `intchain` will use `testnet` (`~/.intchain/testnet` on Linux). This means that attaching to a running testnet node requires the use of a custom endpoint(`intchain attach <datadir>/testnet/intchain.ipc`).
   * The client will connect to the test network, which uses different bootnodes, different network IDs and genesis file.

### Private network

Starting your own private network is easy with init command method.

#### Init private genesis state

First, you need to create the genesis state of your network, which all nodes need to be award of and agree upon.

Init int genesis file and genesis validator
```bash
./bin/intchain --datadir <some directory> init_int_genesis "{799600000000000000000000000, 400000000000000000000000}"
```

Init genesis file

```bash
./bin/intchain --datadir <some directory> init <datadir>/intchain/int_genesis.json
```


#### Starting up node

Starting up a validator node with the genesis state.

```bash
./bin/intchain --datadir <some directory>
```

Starting up a full node with the same genesis state without priv_validator.json file.

```bash
./bin/intchain --datadir <other custom data directory> --bootnodes=<validator node bootnode url above>
```

## Validator Node Based On Testnet

### Create a Wallet

You can create a new wallet, then get some INT from the exchanges or anywhere else into the wallet you just created, .e.g.

```bash
# create a new wallet
intchain account new
```


### Create BLS keys

Once you generate a private validator, it will create a json file priv_validator.json under datadir and restart intchain.

```bash
intchain gen_priv_validator <address>
```

### Confirm your node has caught-up

```bash
intchain attach <datadir>/intchain/intchain.ipc
>int.blockNumber
```


### Become a Candidate

INT Chain is a blockchain system based on IPBFT consensus mechanism, which requires regular replacement of validators to ensure system security.

Epoch is the update cycle of the validator, which is about 2 hours.

You can apply candidate to become a candidate.

1、Apply Candidate

Parameters

   * `from`: address, 32 Bytes - the address which generates the private validator before
   * `securityDeposit`: hex string - amount of security deposit INT(minimum `0x3635c9adc5dea00000`), if you want to be a validator, but there is nobody can vote for you, you should deposit at least `0x152d02c7e14af6800000` INT
   * `commission`: interger - the commission fee percentage (between 0 ~ 100) of each block reward be charged from delegator, when candiate become a validator

Returns

   * `data`: hex string, 32 Bytes - the transaction hash

```bash
# applyCandidate
curl -X POST --data '{"jsonrpc":"2.0","method":"del_applyCandidate","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", "0x152d02c7e14af6800000", 10],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```


### Become a Validator

To become a validator, you can participate by voting. 

#### Vote

1、Get Vote Hash

Parameters

   * `from`: address, 32 Bytes - the address which apply candidate before
   * `publicKey`: hex string, 128 Bytes - bls public key in priv_validator.json
   * `amount`: hex string, the amount of your total vote
   * `salt`: string, random string

Returns

   * `voteHash`: hex string, 32 Bytes - hash of the vote

```bash
# getVoteHash
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_getVoteHash","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", "80DE91BD50F2C3E9F58780C7B739A53A9646DE04BD253D63ED584939B9DB0B381570CAA540D9A84EB5E709FE1D6F80D7AA616990F2177FE3A1F62468A0A26A8809D91C40016C22FA28BE3313F8118C6F4910C95F589980605B23295C9F3CD1FE53BD62F4774B6D29EC16DB98AF830A6E18DA8D1B68331B1C0DA5646FFF359A15", "0x54b40b1f852bda000000", "int"],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```

2、Vote Next Epoch

Parameters

   * `from`: address, 32 Bytes - the address which apply candidate before
   * `voteHash`: hex string, 32 Bytes - the return value of `getVoteHash` above

Returns

   * `data`: hex string, 32 Bytes - the transaction hash 

```bash
# voteNextEpoch
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_voteNextEpoch","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", {vote hash}],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```


#### Reveal Vote

1、Sign Address

Parameters

   * `form`: address, 32 Bytes - the address which apply candidate before
   * `privateKey`: hex string, 32 Bytes - the bls private key in priv_validator.json, you should add `0x` prefix
   
Returns

   * `signature`, 64 Bytes - the bls signature for the address

```bash
# signAddress
curl -X POST --data '{"jsonrpc":"2.0","method":"chain_signAddress","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", "0xF1FBF3781FFDB1A80FA69DB034F4D25CFE27916983B172DC5F5E76384EE7BB2A"],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```

2、Reveal Vote

Parameters

   * `from`: address, 32 Bytes - the address which apply candidate before
   * `pubkey`: hex string, 128 Bytes - the bls public key in priv_validator.json
   * `amount`: hex string, the amount of your total vote
   * `salt`: string, the random string which you have used for `getVoteHash`
   * `signature`: hex string, 64 Bytes - the bls signature of `from` address, signed by bls private key
   
Returns

   * `data`: hex string, 32 Bytes - the transaction hash

```bash
# revealVote
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_revealVote","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", "80DE91BD50F2C3E9F58780C7B739A53A9646DE04BD253D63ED584939B9DB0B381570CAA540D9A84EB5E709FE1D6F80D7AA616990F2177FE3A1F62468A0A26A8809D91C40016C22FA28BE3313F8118C6F4910C95F589980605B23295C9F3CD1FE53BD62F4774B6D29EC16DB98AF830A6E18DA8D1B68331B1C0DA5646FFF359A15", "0x152D02C7E14AF6800000", "int", { signature }],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```

#### Note
Then, if your vote is in the top 100 validators (the max validator size will increase to 100), you will be the validator in the next epoch and you should restart your own node

## Resources
    
   * Explorer: <http://titansexplorer.intchain.io/#/>
   * Wallet: <http://titanswallet.intchain.io/#/>


## Contribution

 See the [contribution](./CONTRIBUTING.md)


## License

[License](./COPYING)



















