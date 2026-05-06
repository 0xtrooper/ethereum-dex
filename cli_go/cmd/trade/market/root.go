package market

import (
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market",
		Short: "Manage markets",
	}

	cmd.AddCommand(newCreateCommand(cfg, ks))
	cmd.AddCommand(newGetCommand(cfg))

	return cmd
}
