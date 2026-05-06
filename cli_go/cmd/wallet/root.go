package wallet

import (
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(svc *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage wallets",
	}

	cmd.AddCommand(newWalletBalanceCommand(svc, ks))
	cmd.AddCommand(newWalletListCommand(ks))
	cmd.AddCommand(newWalletImportCommand(ks))
	cmd.AddCommand(newWalletCreateCommand(ks))
	cmd.AddCommand(newWalletExportCommand(ks))
	cmd.AddCommand(newWalletDeleteCommand(ks))
	cmd.AddCommand(newWalletTerminateCommand(ks))

	return cmd
}
