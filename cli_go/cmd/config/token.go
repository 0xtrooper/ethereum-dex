// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const configTokenRPCConnectTimeout = 10 * time.Second

func newTokenCommand(svc *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage tracked ERC20 tokens",
	}

	cmd.AddCommand(newTokenAddCommand(svc))
	cmd.AddCommand(newTokenListCommand(svc))
	return cmd
}

func newTokenAddCommand(svc *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <address> [symbol]",
		Short: "Add a token to tracked wallet balance tokens",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if !common.IsHexAddress(address) {
				return fmt.Errorf("invalid token address %q", address)
			}
			address = common.HexToAddress(address).Hex()

			symbol := ""
			if len(args) == 2 {
				symbol = strings.TrimSpace(args[1])
			}
			decimals, err := cmd.Flags().GetUint8("decimals")
			if err != nil {
				return err
			}
			hasDecimals := cmd.Flags().Changed("decimals")

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}

			resolvedSymbol := symbol
			resolvedDecimals := decimals
			if resolvedSymbol == "" || !hasDecimals {
				if strings.TrimSpace(cfg.Network.RPCURL) != "" && cfg.Network.ChainID > 0 {
					rpcService, err := service.NewRPC(cfg.Network, configTokenRPCConnectTimeout)
					if err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to connect RPC for token metadata lookup: %v\n", err)
					} else {
						defer rpcService.Close()
						meta, err := svc.ResolveTokenMetadata(context.Background(), rpcService, cfg.Network.ChainID, common.HexToAddress(address))
						if err != nil {
							fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to read token metadata from RPC: %v\n", err)
						} else {
							if resolvedSymbol == "" && strings.TrimSpace(meta.Symbol) != "" {
								resolvedSymbol = strings.TrimSpace(meta.Symbol)
							}
							if !hasDecimals && meta.Decimals > 0 {
								resolvedDecimals = meta.Decimals
								hasDecimals = true
							}
						}
					}
				} else {
					fmt.Fprintln(os.Stderr, "Warning: RPC is not configured, cannot auto-fill missing token metadata.")
				}
			}

			updated := false
			for i := range cfg.Tokens {
				if strings.EqualFold(strings.TrimSpace(cfg.Tokens[i].Address), address) {
					if resolvedSymbol != "" {
						cfg.Tokens[i].Symbol = resolvedSymbol
					}
					if hasDecimals {
						cfg.Tokens[i].Decimals = resolvedDecimals
					}
					updated = true
					break
				}
			}
			if !updated {
				cfg.Tokens = append(cfg.Tokens, service.TokenConfig{
					Symbol:   resolvedSymbol,
					Address:  address,
					Decimals: resolvedDecimals,
				})
			}

			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			tokenRef := service.FormatTokenRef(resolvedSymbol, address)
			if updated {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated token: %s", tokenRef)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Added token: %s", tokenRef)
			}
			if hasDecimals {
				fmt.Fprintf(cmd.OutOrStdout(), " decimals=%d", resolvedDecimals)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Flags().Uint8("decimals", 0, "Token decimals (optional local cache)")
	return cmd
}

func newTokenListCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tokens configured via `config token add`",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, exists, err := svc.LoadIfExists()
			if err != nil {
				return err
			}
			if !exists || len(cfg.Tokens) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No configured tokens.")
				return nil
			}
			for i, t := range cfg.Tokens {
				symbol := strings.TrimSpace(t.Symbol)
				decimals := "-"
				if t.Decimals > 0 {
					decimals = fmt.Sprintf("%d", t.Decimals)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s decimals=%s\n", i+1, service.FormatTokenRef(symbol, common.HexToAddress(t.Address).Hex()), decimals)
			}
			return nil
		},
	}
}
