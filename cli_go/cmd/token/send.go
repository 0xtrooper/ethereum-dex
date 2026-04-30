package token

import (
	"context"
	"fmt"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newSendCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send ERC20 tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenAddress, err := tokenReadAddressFlag(cmd, "token", "Token contract address")
			if err != nil {
				return err
			}

			toAddress, err := tokenReadAddressFlag(cmd, "to", "Recipient address")
			if err != nil {
				return err
			}

			amount, err := tokenReadPositiveBigIntFlag(cmd, "amount", "Send amount")
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

			tx, err := tokenService.Send(context.Background(), toAddress, amount)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", walletService.Address())
			fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", tokenAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Recipient: %s\n", toAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Amount: %s\n", amount.String())
			fmt.Fprintf(cmd.OutOrStdout(), "Send tx: %s\n", tx.Hash().Hex())
			return nil
		},
	}

	cmd.Flags().String("token", "", "Token contract address")
	cmd.Flags().String("to", "", "Recipient address")
	cmd.Flags().String("amount", "", "Token amount (integer, wei-style units)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	return cmd
}
