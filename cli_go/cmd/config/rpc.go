package config

import (
	"fmt"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newRPCCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "rpc <url>",
		Short: "Set the Ethereum RPC URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rpcURL := strings.TrimSpace(args[0])
			if rpcURL == "" {
				return fmt.Errorf("rpc url cannot be empty")
			}

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}
			cfg.Network.RPCURL = rpcURL
			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set rpc_url: %s\n", cfg.Network.RPCURL)
			return nil
		},
	}
}
