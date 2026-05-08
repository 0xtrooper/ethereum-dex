# Serverless Ethereum DEX

This is a fully on-chain decentralized limit order book on Ethereum. The system is serverless and can be synced using a local Ethereum RPC. It has a risk isolation system to prevent malicious tokens from attacking other markets, and it has support for both regular ERC20 and non-standard fee-for-transfer tokens.

# Setting up a local node

The contracts are complete and can be compiled using the Solidity compiler available [here](https://docs.soliditylang.org/en/latest/installing-solidity.html).

To compile the CLI:

```
cd cli_go
go build
```
