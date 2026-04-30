package contact

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
		Use:   "contract <address>",
		Short: "Set the DEX contract address",
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
