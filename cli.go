package main

import (
	"context"
	"flag"
	"log"	
	"os"
	"strconv"
	"math/big"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

)

var ABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"orderId","type":"uint256"}],"name":"OrderCanceled","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"orderId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"baseQuantity","type":"uint256"}],"name":"OrderFill","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"orderId","type":"uint256"},{"indexed":true,"internalType":"address","name":"user","type":"address"},{"indexed":true,"internalType":"address","name":"baseToken","type":"address"},{"indexed":true,"internalType":"address","name":"quoteToken","type":"address"},{"indexed":false,"internalType":"enum OrderBook.Side","name":"side","type":"uint8"},{"indexed":false,"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"quoteQuantity","type":"uint256"}],"name":"OrderPlaced","type":"event"},{"inputs":[{"internalType":"uint256","name":"orderId","type":"uint256"}],"name":"cancelOrder","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"orderId","type":"uint256"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"}],"name":"fillOrder","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256[]","name":"orderIds","type":"uint256[]"},{"internalType":"uint256[]","name":"baseFillQuantities","type":"uint256[]"}],"name":"fillOrders","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"orderCounter","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"orders","outputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"internalType":"uint256","name":"quoteQuantity","type":"uint256"},{"internalType":"uint256","name":"orderId","type":"uint256"},{"internalType":"address","name":"baseToken","type":"address"},{"internalType":"address","name":"quoteToken","type":"address"},{"internalType":"enum OrderBook.Side","name":"side","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"baseToken","type":"address"},{"internalType":"address","name":"quoteToken","type":"address"},{"internalType":"enum OrderBook.Side","name":"side","type":"uint8"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"internalType":"uint256","name":"quoteQuantity","type":"uint256"}],"name":"placeOrder","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

// Test Key
var SK   = "0xaf5ead4413ff4b78bc94191a2926ae9ccbec86ce099d65aaf469e9eb1a0fa87f"
var ADDR = "0x6177843db3138ae69679A54b95cf345ED759450d"

var CONTRACT_ADDRESS = common.HexToAddress("0x00")

func main() {
	StorageMetaData := bind.MetaData{ ABI: ABI }
	ParsedABI, err := StorageMetaData.ParseABI()
	if err != nil {
		log.Println("error parsing ABI", err)
		os.Exit(1)
	}

	rpcurl := flag.String("rpcurl", "http://localhost:8545", "Ethereum RPC URL")	
	flag.Parse()

	command := os.Args[1]

	switch command {
		case "placeorder":
			baseToken := os.Args[2]
			quoteToken := os.Args[3]
			side := os.Args[4]
			price, err := strconv.ParseFloat(os.Args[5], 64)
			if err != nil {
				log.Println("error parsing price", err)
				os.Exit(1)
			}
			quantity, err := strconv.ParseFloat(os.Args[6], 64)
			if err != nil {
				log.Println("error parsing quantity", err)
				os.Exit(1)
			}
			if side != "buy" && side != "sell" {
				log.Println("side must be 'buy' or 'sell'")
				os.Exit(1)
			}

			conn, err := ethclient.Dial(*rpcurl)
			if err != nil {
				log.Println("Failed to connect to rpc: ", err)
				os.Exit(1)
			}
			
			packedTxData, err := ParsedABI.Pack("TODO")
			if err != nil {
				log.Println("Error packing data", err)
				os.Exit(1)
			}

			err = sendTransaction(conn, SK, packedTxData)
			

			
		case "fillorder":
			orderId := os.Args[2]
			quantity, err := strconv.ParseFloat(os.Args[3], 64)
			if err != nil {
				log.Println("error parsing quantity", err)
				os.Exit(1)
			}

			conn, err := ethclient.Dial(*rpcurl)
			if err != nil {
				log.Println("Failed to connect to rpc: ", err)
				os.Exit(1)
			}
			
			packedTxData, err := ParsedABI.Pack("TODO")
			if err != nil {
				log.Println("Error packing data", err)
				os.Exit(1)
			}

			err = sendTransaction(conn, SK, packedTxData)

		case "cancelorder":
			orderId := os.Args[2]

			conn, err := ethclient.Dial(*rpcurl)
			if err != nil {
				log.Println("Failed to connect to rpc: ", err)
				os.Exit(1)
			}
			
			packedTxData, err := ParsedABI.Pack("TODO")
			if err != nil {
				log.Println("Error packing data", err)
				os.Exit(1)
			}

			err = sendTransaction(conn, SK, packedTxData)
		case "sync":
		case "test":
			log.Println("Running test cases")
		case "help":
		default:
			log.Println("Long HELP message TBD")
	}
}

// sendTransaction sends a transaction with 1 ETH to a specified address.
func sendTransaction(cl *ethclient.Client, privkeyhex string, txdata []byte) error {
	sk, err := crypto.ToECDSA(common.FromHex(privkeyhex))
	if err != nil {
		return err
	}
	sender := crypto.PubkeyToAddress(sk.PublicKey)

	// Retrieve the chainid (needed for signer)
	chainid, err := cl.ChainID(context.Background())
	if err != nil {
		return err
	}

	// Retrieve the pending nonce
	nonce, err := cl.PendingNonceAt(context.Background(), sender)
	if err != nil {
		return err
	}

	// Get suggested gas price
	tipCap, err := cl.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}
	feeCap, err := cl.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	// Create a new transaction
	tx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   chainid,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			To:        &CONTRACT_ADDRESS,
			Value:     big.NewInt(0),
			Data:      txdata,
		})
	// Sign the transaction using our keys
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
	if err != nil {
		return err
	}

	// Send the transaction to our node
	return cl.SendTransaction(context.Background(), signedTx)
}
