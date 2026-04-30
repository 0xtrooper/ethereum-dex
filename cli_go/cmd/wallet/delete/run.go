package walletdelete

import "github.com/spf13/cobra"

type WalletDeleter interface {
	List() []string
	Delete(address string) error
}

func NewCommand(ks WalletDeleter) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [address...]",
		Short: "Delete one or more wallets from the keystore",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
}

func Run(cmd *cobra.Command, ks WalletDeleter, args []string) error {
	wallets := ks.List()
	in, err := input(cmd, args, wallets)
	if err != nil {
		return err
	}
	out, err := process(ks, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
