// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package market

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func marketReadContractAddress(cmd *cobra.Command, cfg *service.Config) (string, error) {
	contractAddress, _ := cmd.Flags().GetString("contract")
	contractAddress = strings.TrimSpace(contractAddress)
	if contractAddress == "" && cfg != nil {
		c := strings.TrimSpace(cfg.Contract.Address)
		if common.IsHexAddress(c) && common.HexToAddress(c) != (common.Address{}) {
			contractAddress = c
		}
	}
	for contractAddress == "" {
		next, err := marketAsk("OrderBook contract address")
		if err != nil {
			return "", err
		}
		contractAddress = strings.TrimSpace(next)
	}
	if !common.IsHexAddress(contractAddress) {
		return "", fmt.Errorf("invalid contract address %q", contractAddress)
	}
	if common.HexToAddress(contractAddress) == (common.Address{}) {
		return "", fmt.Errorf("contract address cannot be zero address")
	}
	return contractAddress, nil
}

func marketReadAddressFlag(cmd *cobra.Command, flagName string, promptLabel string) (common.Address, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := marketAsk(promptLabel)
			if err != nil {
				return common.Address{}, err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return common.Address{}, err
			}
		}
		if !common.IsHexAddress(value) {
			fmt.Fprintf(os.Stderr, "Invalid address %q.\n", value)
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return common.Address{}, err
			}
			continue
		}
		return common.HexToAddress(value), nil
	}
}

func marketAsk(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("aborted")
	}
	return scanner.Text(), nil
}

func marketSelectWalletAddress(ks *service.Keystore, requestedAddress string) (string, error) {
	if ks == nil {
		return "", fmt.Errorf("keystore service is not initialized")
	}
	wallets := ks.List()
	if len(wallets) == 0 {
		return "", fmt.Errorf("no wallets in keystore")
	}

	if strings.TrimSpace(requestedAddress) != "" {
		if !common.IsHexAddress(requestedAddress) {
			return "", fmt.Errorf("invalid wallet address %q", requestedAddress)
		}
		wanted := common.HexToAddress(requestedAddress).Hex()
		for _, wallet := range wallets {
			if strings.EqualFold(wallet, wanted) {
				return wanted, nil
			}
		}
		return "", fmt.Errorf("wallet %s not found in keystore", wanted)
	}

	if len(wallets) == 1 {
		return wallets[0], nil
	}

	for i, wallet := range wallets {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, wallet)
	}
	fmt.Fprintln(os.Stderr)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "Select wallet [1-%d]: ", len(wallets))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("aborted")
		}
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(wallets) {
			fmt.Fprintf(os.Stderr, "Please enter a number between 1 and %d.\n", len(wallets))
			continue
		}
		return wallets[n-1], nil
	}
}
