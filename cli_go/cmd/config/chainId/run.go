package chainid

import (
	"dex/service"

	"github.com/spf13/cobra"
)

type ConfigService interface {
	Ensure() (*service.Config, bool, error)
	Save(*service.Config) error
	Path() string
}

func NewCommand(svc ConfigService) *cobra.Command {
	return &cobra.Command{
		Use:   "chain-id <id>",
		Short: "Set the Ethereum chain ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, svc, args)
		},
	}
}

func Run(cmd *cobra.Command, svc ConfigService, args []string) error {
	in, err := input(cmd, args)
	if err != nil {
		return err
	}
	out, err := process(svc, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
