package wallet

import (
	"fmt"

	"dex/service"

	"github.com/spf13/cobra"
)

func newWalletListCommand(ks *service.Keystore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all wallets in the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			addresses := ks.List()
			if len(addresses) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No wallets in keystore.")
				return nil
			}
			for i, addr := range addresses {
				fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
			}
			return nil
		},
	}
}
