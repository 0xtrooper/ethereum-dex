package walletterminate

import "github.com/spf13/cobra"

type WalletTerminator interface {
	List() []string
	DeleteAll() error
}

func NewCommand(ks WalletTerminator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Delete all wallets from the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func Run(cmd *cobra.Command, ks WalletTerminator, args []string) error {
	wallets := ks.List()
	in, err := input(cmd, wallets)
	if err != nil {
		return err
	}
	out, err := process(ks, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
