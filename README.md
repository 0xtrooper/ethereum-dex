# Serverless Ethereum DEX

This is a fully on-chain decentralized limit order book on Ethereum. The system is serverless and can be synced using a local Ethereum RPC. It has a risk isolation system to prevent malicious tokens from attacking other markets, and it has support for both regular ERC20 and non-standard fee-for-transfer tokens.

# Setting up a local node

To compile the CLI:

```
cd cli_go
go build
```

A binary called `dex` will be available in the `cli_go` folder.

# CLI Commands

### Help

```
$ ./dex help
Decentralized Exchange Command Line Interface.

Usage:
  dex [command]

Available Commands:
  config      Manage dex configuration
  help        Help about any command
  token       Interact with ERC20 tokens
  trade       Interact with the exchange order book
  wallet      Manage wallets

Flags:
  -h, --help   help for dex

Use "dex [command] --help" for more information about a command.
```

### Setting the Chain ID

By default the Chain ID is set to Ethereum mainnet (chainId: 1). To change it to a different network (example: arbitrum):

```
$ ./dex config chain-id 42161
Set chain_id: 42161
```

### Setting the Contract Address

In the future, we will have default contract addresses stored for many chains. 

For now, contract addresses have to be set manually. 

```
./dex config contract 0xc026deC188ef7D8B4742CE38E36f3C3BcD328E4C
```

### Setting an RPC

The RPC will default to http://localhost:8545 (WIP - for now set manually). To use a remote RPC: 

```
./dex config rpc https://rpc.flashbots.net
```

Unless you are using a local node, you will likely need a paid tier RPC to sync orderbooks.

### Create a Wallet

```
./dex wallet create
```

Follow the help text in the interactive prompts to set up a wallet.

### Other Wallet Commands

```
$ ./dex help wallet
Manage wallets

Usage:
  dex wallet [command]

Available Commands:
  balance     Show non-zero token balances for tracked tokens
  create      Create a new wallet
  delete      Delete one or more wallets from the keystore
  export      Export a wallet private key
  import      Import a wallet into the keystore
  list        List all wallets in the keystore
  terminate   Delete all wallets from the keystore

Flags:
  -h, --help   help for wallet

Use "dex wallet [command] --help" for more information about a command.
```

### Configuration

```
$ ./dex help config
Manage dex configuration

Usage:
  dex config [command]

Available Commands:
  chain-id    Set the Ethereum chain ID
  contract    Set the DEX contract address
  rpc         Set the Ethereum RPC URL
  show        Show dex configuration
  terminate   Delete the dex configuration file from disk
  token       Manage tracked ERC20 tokens

Flags:
  -h, --help   help for config

Use "dex config [command] --help" for more information about a command.
```

### Trading

```
$ ./dex trade order
Manage orders

Usage:
  dex trade order [command]

Available Commands:
  cancel      Cancel an order you own
  count       Get total order counter
  fill        Fill an existing order
  get         Get a specific order
  place       Place a new order

Flags:
  -h, --help   help for order

Use "dex trade order [command] --help" for more information about a command.
```
