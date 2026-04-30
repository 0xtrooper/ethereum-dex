package config

import (
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(svc *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage dex configuration",
	}

	cmd.AddCommand(newRPCCommand(svc))
	cmd.AddCommand(newChainIDCommand(svc))
	cmd.AddCommand(newContractCommand(svc))
	cmd.AddCommand(newShowCommand(svc))
	cmd.AddCommand(newTerminateCommand(svc))

	return cmd
}
