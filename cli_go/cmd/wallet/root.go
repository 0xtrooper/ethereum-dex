package wallet

import (
	createcmd "dex/cmd/wallet/create"
	deletecmd "dex/cmd/wallet/delete"
	exportcmd "dex/cmd/wallet/export"
	importcmd "dex/cmd/wallet/import"
	listcmd "dex/cmd/wallet/list"
	terminatecmd "dex/cmd/wallet/terminate"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage wallets",
	}

	cmd.AddCommand(listcmd.NewCommand(ks))
	cmd.AddCommand(importcmd.NewCommand(ks))
	cmd.AddCommand(createcmd.NewCommand(ks))
	cmd.AddCommand(exportcmd.NewCommand(ks))
	cmd.AddCommand(deletecmd.NewCommand(ks))
	cmd.AddCommand(terminatecmd.NewCommand(ks))

	return cmd
}
