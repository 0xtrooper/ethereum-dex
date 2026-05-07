package order

import (
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Manage orders",
	}

	cmd.AddCommand(newPlaceCommand(cfg, ks))
	cmd.AddCommand(newGetCommand(cfg))
	cmd.AddCommand(newCountCommand(cfg))
	cmd.AddCommand(newFillCommand(cfg, ks))
	cmd.AddCommand(newCancelCommand(cfg, ks))

	return cmd
}
