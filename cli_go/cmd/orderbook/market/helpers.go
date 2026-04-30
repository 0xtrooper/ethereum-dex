package market

import (
	"bufio"
	"fmt"
	"os"
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
