package main

import (
	"context"
	"flag"
	"fmt"
	"log"	
	"os"
	"strconv"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"

)

var ABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"orderId","type":"uint256"}],"name":"OrderCanceled","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"orderId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"baseQuantity","type":"uint256"}],"name":"OrderFill","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"orderId","type":"uint256"},{"indexed":true,"internalType":"address","name":"user","type":"address"},{"indexed":true,"internalType":"address","name":"baseToken","type":"address"},{"indexed":true,"internalType":"address","name":"quoteToken","type":"address"},{"indexed":false,"internalType":"enum OrderBook.Side","name":"side","type":"uint8"},{"indexed":false,"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"quoteQuantity","type":"uint256"}],"name":"OrderPlaced","type":"event"},{"inputs":[{"internalType":"uint256","name":"orderId","type":"uint256"}],"name":"cancelOrder","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"orderId","type":"uint256"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"}],"name":"fillOrder","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256[]","name":"orderIds","type":"uint256[]"},{"internalType":"uint256[]","name":"baseFillQuantities","type":"uint256[]"}],"name":"fillOrders","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"orderCounter","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"orders","outputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"internalType":"uint256","name":"quoteQuantity","type":"uint256"},{"internalType":"uint256","name":"orderId","type":"uint256"},{"internalType":"address","name":"baseToken","type":"address"},{"internalType":"address","name":"quoteToken","type":"address"},{"internalType":"enum OrderBook.Side","name":"side","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"baseToken","type":"address"},{"internalType":"address","name":"quoteToken","type":"address"},{"internalType":"enum OrderBook.Side","name":"side","type":"uint8"},{"internalType":"uint256","name":"baseQuantity","type":"uint256"},{"internalType":"uint256","name":"quoteQuantity","type":"uint256"}],"name":"placeOrder","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
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

			chainID, err := conn.ChainID(context.Background())
			if err != nil {
				panic(fmt.Errorf("failed to retrieve chain ID: %v", err))
			}

			OrderBookContract := bind.NewBoundContract(CONTRACT_ADDRESS, *ParsedABI, conn, conn, conn)

			packedData, err := ParsedABI.Pack("TODO")
			if err != nil {
				log.Println("Error packing data", err)
				os.Exit(1)
			}
			//key, err := keystore.DecryptKey(json, passphrase)
			//auth := bind.NewKeyedTransactor(key.PrivateKey, chainID)

			transactOpts := bind.TransactOpts{} // TODO: Fill in
			val, err := bind.Transact(OrderBookContract, &transactOpts, packedData)
			

			
		case "fillorder":
			orderId := os.Args[2]
			quantity, err := strconv.ParseFloat(os.Args[3], 64)
			if err != nil {
				log.Println("error parsing quantity", err)
				os.Exit(1)
			}
		case "cancelorder":
			orderId := os.Args[2]
		case "sync":
			baseToken := os.Args[2]
			quoteToken := os.Args[3]
		case "test":
			log.Println("Running test cases")
		case "help":
		default:
			log.Println("Long HELP message TBD")
	}
}
