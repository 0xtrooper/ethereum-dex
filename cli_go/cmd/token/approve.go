package token

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const tokenRPCConnectTimeout = 15 * time.Second

func newApproveCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve token allowance for a spender",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenAddress, err := tokenReadAddressFlag(cmd, "token", "Token contract address")
			if err != nil {
				return err
			}

			spenderAddress, usedDefaultSpender, err := tokenReadSpenderAddress(cmd, cfg.Get())
			if err != nil {
				return err
			}
			if usedDefaultSpender {
				fmt.Fprintf(cmd.OutOrStdout(), "No spender was provided. Using default exchange contract from config: %s\n", spenderAddress.Hex())
			}

			amount, err := tokenReadPositiveBigIntFlag(cmd, "amount", "Approve amount")
			if err != nil {
				return err
			}

			walletAddress, _ := cmd.Flags().GetString("wallet")
			walletAddress = strings.TrimSpace(walletAddress)

			rpcService, err := service.NewRPC(cfg.Get().Network, tokenRPCConnectTimeout)
			if err != nil {
				return err
			}
			defer rpcService.Close()

			walletService, err := service.NewWallet(ks, walletAddress)
			if err != nil {
				return err
			}

			tokenService, err := service.NewTokenService(rpcService, walletService, tokenAddress, cfg.Get().Network.ChainID)
			if err != nil {
				return err
			}

			tx, err := tokenService.Approve(context.Background(), spenderAddress, amount)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", walletService.Address())
			fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", tokenAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Spender: %s\n", spenderAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Amount: %s\n", amount.String())
			fmt.Fprintf(cmd.OutOrStdout(), "Approve tx: %s\n", tx.Hash().Hex())
			return nil
		},
	}

	cmd.Flags().String("token", "", "Token contract address")
	cmd.Flags().String("spender", "", "Spender address (defaults to config.contract.address)")
	cmd.Flags().String("amount", "", "Allowance amount (integer, wei-style units)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	return cmd
}

func tokenReadSpenderAddress(cmd *cobra.Command, cfg *service.Config) (common.Address, bool, error) {
	spenderValue, _ := cmd.Flags().GetString("spender")
	spenderValue = strings.TrimSpace(spenderValue)
	if spenderValue != "" {
		if !common.IsHexAddress(spenderValue) {
			return common.Address{}, false, fmt.Errorf("invalid spender address %q", spenderValue)
		}
		return common.HexToAddress(spenderValue), false, nil
	}

	if cfg != nil {
		configAddress := strings.TrimSpace(cfg.Contract.Address)
		if common.IsHexAddress(configAddress) {
			spender := common.HexToAddress(configAddress)
			if spender != (common.Address{}) {
				return spender, true, nil
			}
		}
	}

	spender, err := tokenReadAddressPrompt("Spender address")
	if err != nil {
		return common.Address{}, false, err
	}
	return spender, false, nil
}

func tokenReadAddressFlag(cmd *cobra.Command, flagName string, promptLabel string) (common.Address, error) {
	value, _ := cmd.Flags().GetString(flagName)
	value = strings.TrimSpace(value)
	if value == "" {
		return tokenReadAddressPrompt(promptLabel)
	}
	if !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("invalid %s %q", flagName, value)
	}
	return common.HexToAddress(value), nil
}

func tokenReadAddressPrompt(label string) (common.Address, error) {
	for {
		value, err := tokenAsk(label)
		if err != nil {
			return common.Address{}, err
		}
		value = strings.TrimSpace(value)
		if !common.IsHexAddress(value) {
			fmt.Fprintf(os.Stderr, "Invalid address %q.\n", value)
			continue
		}
		return common.HexToAddress(value), nil
	}
}

func tokenReadPositiveBigIntFlag(cmd *cobra.Command, flagName string, promptLabel string) (*big.Int, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := tokenAsk(promptLabel)
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
		if n.Sign() <= 0 {
			fmt.Fprintln(os.Stderr, "Amount must be greater than zero.")
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		return n, nil
	}
}

func tokenAsk(label string) (string, error) {
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
