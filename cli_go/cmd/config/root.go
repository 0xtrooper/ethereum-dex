package config

import (
	chainidcmd "dex/cmd/config/chainId"
	contactcmd "dex/cmd/config/contact"
	rpccmd "dex/cmd/config/rpc"
	showcmd "dex/cmd/config/show"
	terminatecmd "dex/cmd/config/terminate"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(svc *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage dex configuration",
	}

	cmd.AddCommand(rpccmd.NewCommand(svc))
	cmd.AddCommand(chainidcmd.NewCommand(svc))
	cmd.AddCommand(contactcmd.NewCommand(svc))
	cmd.AddCommand(showcmd.NewCommand(svc))
	cmd.AddCommand(terminatecmd.NewCommand(svc))

	return cmd
}
