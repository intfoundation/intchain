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


```bash
# applyCandidate
curl -X POST --data '{"jsonrpc":"2.0","method":"del_applyCandidate","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", "0x152d02c7e14af6800000", 10],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8555/intchain
```


### Become a Validator

To become a validator, you can participate by voteing. 

#### Vote

```bash
# getVoteHash return vote hash
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_getVoteHash","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", {your bls public key in priv_validator.json}, "0x54b40b1f852bda000000", "int"],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```

```bash
# voteNextEpoch
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_voteNextEpoch","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", {vote hash}],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```


#### Reveal Vote

```bash
# signAddress return sign hash
curl -X POST --data '{"jsonrpc":"2.0","method":"chain_signAddress","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", {your bls private key in priv_validator.json}],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```

```bash
# revealVote
curl -X POST --data '{"jsonrpc":"2.0","method":"tdm_revealVote","params":["INT3LJK4UctyCwv5mdvnpBYvMbRTZBia", {your bls public key in priv_validator.json}, "0x152D02C7E14AF6800000", "int", {sign hash}],"id":1}' -H 'content-type: application/json;' http://127.0.0.1:8556/testnet
```


## Resources
    
   * Explorer: <http://titansexplorer.intchain.io/#/>
   * Wallet: <http://titanswallet.intchain.io/#/>


## Contribution

 See the [contribution](./CONTRIBUTING.md)


## License

[License](./COPYING)



















