package config

import (
	"fmt"
	"regexp"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

var ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

func newContractCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "contract <address>",
		Short: "Set the DEX contract address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if !ethAddressPattern.MatchString(address) {
				return fmt.Errorf("invalid Ethereum address %q", address)
			}

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}
			cfg.Contract.Address = address
			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set contract address: %s\n", cfg.Contract.Address)
			return nil
		},
	}
}
