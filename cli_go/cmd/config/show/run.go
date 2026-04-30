package show

import (
	"dex/service"

	"github.com/spf13/cobra"
)

type ConfigLoader interface {
	LoadIfExists() (*service.Config, bool, error)
}

func NewCommand(svc ConfigLoader) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show dex configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, svc, args)
		},
	}
}

func Run(cmd *cobra.Command, svc ConfigLoader, args []string) error {
	out, err := process(svc)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
