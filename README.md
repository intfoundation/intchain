# INT Chain

![intchain](https://raw.githubusercontent.com/intfoundation/intchain/master/docs/intchain.jpg)

INT Chain is the world's first bottom up new-generation blockchain of things (BoT) communication standard and base application platform. The ecosystem is specifically designed for easy integration with any IoT protocol.

To both improve and encourage device interconnectivity, we have built an economy driven ecosystem by providing token-incentives through a decentralized TCP/IP based architecture of IoT. This new business model, molded by IoT devices, will support an entirely new ecosystem of the Internet of Things.


## Install

### Latest Version

The latest INT Chain version for Testnet is v4.0.4

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
git clone --branch testnet https://github.com/intfoundation/intchain
cd intchain
make intchain
```

If your environment variables have set up correctly, you should not get any errors by running the above commands.
Now check your `intchain` version.

```bash
intchain version
```

## Running `intchain`

`intchain` is the entry point into the INT Chain network(main, test or private network). It can be used by other processes as a gateway into the INT Chain network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports.

We've enumerated a few common command to let you run your own `intchain` instance quickly. If you want to look over all command line options, please use `intchain --help`


### Main network

The most common condition is that users want to simply interact with the INT Chain network: create accounts; transfer funds; deploy and interact with contracts.

```bash
intchain 
```

Then you can attach to an already running `intchain` instance with `intchain attach`, and you can invoke all official methods.

```bash
intchain attach <datadir>/intchain/intchain.ipc
```


### Test network

For developers, you would like to deploy contracts for testing, but you do not want to do that with real money spending.
In other words, instead of attaching to the main network, you want to join the test network with your node, which is fully equivalent to the main network.

```bash
intchain --testnet
```

Specifying the `--testnet` flag, will reconfigure `intchain` instance a bit:
   
   * Instead of using the default data directory(`~/.intchain/intchain` on Linux for example), `intchain` will use `testnet` (`~/.intchain/testnet` on Linux). This means that attaching to a running testnet node requires the use of a custom endpoint(`intchain attach <datadir>/testnet/intchain.ipc`).
   * The client will connect to the test network, which uses different bootnodes, different network IDs and genesis file.

## Resources
    
   * Explorer: <http://titansexplorer.intchain.io/#/>
   * Wallet: <http://titanswallet.intchain.io/#/>
   * Document: <https://titansdocs.intchain.io/>


## Contribution

 See the [contribution](./CONTRIBUTING.md)


## License

[License](./COPYING)



















