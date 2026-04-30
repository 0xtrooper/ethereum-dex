package config

import (
	"fmt"
	"strconv"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newChainIDCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "chain-id <id>",
		Short: "Set the Ethereum chain ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawChainID := strings.TrimSpace(args[0])
			chainID, err := strconv.ParseInt(rawChainID, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain id %q: %w", rawChainID, err)
			}
			if chainID <= 0 {
				return fmt.Errorf("chain id must be greater than zero")
			}

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}
			cfg.Network.ChainID = chainID
			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set chain_id: %d\n", cfg.Network.ChainID)
			return nil
		},
	}
}
