package order

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strings"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func orderReadContractAddress(cmd *cobra.Command, cfg *service.Config) (string, error) {
	contractAddress, _ := cmd.Flags().GetString("contract")
	contractAddress = strings.TrimSpace(contractAddress)
	if contractAddress == "" && cfg != nil {
		configAddress := strings.TrimSpace(cfg.Contract.Address)
		if common.IsHexAddress(configAddress) && common.HexToAddress(configAddress) != (common.Address{}) {
			contractAddress = configAddress
		}
	}
	for contractAddress == "" {
		next, err := orderAsk("OrderBook contract address")
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

func orderReadAddressFlag(cmd *cobra.Command, flagName string, promptLabel string) (common.Address, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := orderAsk(promptLabel)
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

func orderReadBigIntFlag(cmd *cobra.Command, flagName string, promptLabel string, mustBePositive bool) (*big.Int, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			if !mustBePositive {
				return big.NewInt(0), nil
			}
			next, err := orderAsk(promptLabel)
			if err != nil {
				return nil, err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return nil, err
			}
		}

		n, ok := new(big.Int).SetString(value, 10)
		if !ok {
			fmt.Fprintf(os.Stderr, "Invalid integer value %q.\n", value)
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		if mustBePositive && n.Sign() <= 0 {
			fmt.Fprintln(os.Stderr, "Value must be greater than zero.")
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		if !mustBePositive && n.Sign() < 0 {
			fmt.Fprintln(os.Stderr, "Value cannot be negative.")
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		return n, nil
	}
}

func orderReadSide(cmd *cobra.Command) (uint8, error) {
	for {
		value, _ := cmd.Flags().GetString("side")
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			next, err := orderAsk("Order side [buy/sell]")
			if err != nil {
				return 0, err
			}
			value = strings.ToLower(strings.TrimSpace(next))
			if err := cmd.Flags().Set("side", value); err != nil {
				return 0, err
			}
		}
		switch value {
		case "buy", "0":
			return service.OrderSideBuy, nil
		case "sell", "1":
			return service.OrderSideSell, nil
		default:
			fmt.Fprintln(os.Stderr, "Invalid side. Please use buy or sell.")
			if err := cmd.Flags().Set("side", ""); err != nil {
				return 0, err
			}
		}
	}
}

func orderAsk(label string) (string, error) {
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
