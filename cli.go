package main

import (
	"flag"
	"fmt"
	"log"	
	"os"
	"strconv"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func main() {
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
		case "help":
			log.Println("Long HELP message TBD")
		case "test":
			log.Println("Running test cases")
	}
}
