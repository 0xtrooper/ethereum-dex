package configterminate

import "github.com/spf13/cobra"

type ConfigTerminator interface {
	Path() string
	Delete() error
}

func NewCommand(svc ConfigTerminator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Delete the dex configuration file from disk",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, svc, args)
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func Run(cmd *cobra.Command, svc ConfigTerminator, args []string) error {
	in, err := input(cmd, svc)
	if err != nil {
		return err
	}
	out, err := process(svc, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
