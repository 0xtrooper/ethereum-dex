package config

import (
	"fmt"

	"dex/service"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v2"
)

func newShowCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show dex configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, exists, err := svc.LoadIfExists()
			if err != nil {
				return err
			}
			if !exists {
				fmt.Fprintf(cmd.OutOrStdout(), "No config found. Showing defaults.\n\n")
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}
