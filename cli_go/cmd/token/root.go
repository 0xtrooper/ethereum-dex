package token

import (
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Interact with ERC20 tokens",
	}

	cmd.AddCommand(newApproveCommand(cfg, ks))
	cmd.AddCommand(newSendCommand(cfg, ks))

	return cmd
}
