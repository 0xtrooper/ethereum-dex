package walletlist

import "github.com/spf13/cobra"

type WalletLister interface {
	List() []string
}

func NewCommand(ks WalletLister) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all wallets in the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
}

func Run(cmd *cobra.Command, ks WalletLister, args []string) error {
	out, err := process(ks)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
