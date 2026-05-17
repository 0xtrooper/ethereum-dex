// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package wallet

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"dex/internal/amount"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const walletBalanceRPCConnectTimeout = 15 * time.Second

type walletBalanceEntry struct {
	symbol  string
	address string
	raw     *big.Int
	display string
}

type walletBalanceResult struct {
	address string
	entries []walletBalanceEntry
}

func newWalletBalanceCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Show non-zero token balances for tracked tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWalletBalance(cmd, cfg, ks)
		},
	}
	cmd.Flags().String("wallet", "", "Wallet address (if omitted, show balances for all wallets)")
	return cmd
}

func runWalletBalance(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	walletFlag, _ := cmd.Flags().GetString("wallet")
	targets, err := resolveWalletBalanceTargets(strings.TrimSpace(walletFlag), ks)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true

	rpcService, err := service.NewRPC(cfg.Get().Network, walletBalanceRPCConnectTimeout)
	if err != nil {
		return err
	}
	defer rpcService.Close()

	tracked := service.MergeKnownAndConfiguredTokens(cfg.Get().Network.ChainID, cfg.Get().Tokens)
	if len(tracked) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tracked tokens configured for this chain.")
		fmt.Fprintln(cmd.OutOrStdout(), "Add tokens with: dex config token add <token_address> [symbol]")
		return nil
	}

	results := make([]walletBalanceResult, 0, len(targets))
	for _, walletAddr := range targets {
		out, err := walletBalancesForAddress(rpcService, cfg, walletAddr, tracked)
		if err != nil {
			return err
		}
		results = append(results, out)
	}

	printed := 0
	for _, r := range results {
		if len(r.entries) == 0 {
			continue
		}
		printed++
		fmt.Fprintf(cmd.OutOrStdout(), "Wallet: %s\n", r.address)
		for _, e := range r.entries {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s (raw: %s) [%s]\n", e.symbol, e.display, e.raw.String(), e.address)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
	if printed == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No non-zero tracked token balances found.")
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Hint: If you are missing a token, add it with `dex config token add <token_address> [symbol]` or check www.etherscan.io/token.")
	return nil
}

func resolveWalletBalanceTargets(walletFlag string, ks *service.Keystore) ([]string, error) {
	wallets := ks.List()
	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets in keystore")
	}

	if walletFlag == "" {
		return wallets, nil
	}
	if !common.IsHexAddress(walletFlag) {
		return nil, fmt.Errorf("invalid wallet address %q", walletFlag)
	}
	wanted := common.HexToAddress(walletFlag).Hex()
	for _, w := range wallets {
		if strings.EqualFold(w, wanted) {
			return []string{common.HexToAddress(w).Hex()}, nil
		}
	}
	return nil, fmt.Errorf("wallet %s not found in keystore", wanted)
}

func walletBalancesForAddress(rpcService *service.RPC, cfg *service.Service, walletAddress string, tracked []service.ChainToken) (walletBalanceResult, error) {
	ctx := context.Background()
	owner := common.HexToAddress(walletAddress)
	out := walletBalanceResult{
		address: owner.Hex(),
		entries: make([]walletBalanceEntry, 0),
	}

	for _, token := range tracked {
		if !common.IsHexAddress(token.Address) {
			continue
		}
		tokenAddress := common.HexToAddress(token.Address)
		meta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, tokenAddress)
		if err == nil {
			token = meta
		}

		tokenService, err := service.NewTokenService(rpcService, nil, tokenAddress, cfg.Get().Network.ChainID)
		if err != nil {
			continue
		}

		balance, err := tokenService.BalanceOf(ctx, owner)
		if err != nil || balance.Sign() == 0 {
			continue
		}

		symbol := strings.TrimSpace(token.Symbol)
		if symbol == "" {
			symbol = "TOKEN"
		}

		decimals := token.Decimals
		if decimals == 0 {
			decimals = 18
		}

		out.entries = append(out.entries, walletBalanceEntry{
			symbol:  symbol,
			address: tokenAddress.Hex(),
			raw:     balance,
			display: amount.FormatUnits(balance, decimals),
		})
	}

	return out, nil
}
