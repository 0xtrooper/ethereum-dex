package walletcreate

import "github.com/spf13/cobra"

type WalletCreator interface {
	Create(password string) (string, error)
	GenerateMnemonic() (string, error)
	CreateFromMnemonic(mnemonic string, password string) (string, error)
}

func NewCommand(ks WalletCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
	cmd.Flags().String("password", "", "Keystore password")
	cmd.Flags().Bool("no-mnemonic", false, "Skip mnemonic phrase generation and confirmation")
	return cmd
}

func Run(cmd *cobra.Command, ks WalletCreator, args []string) error {
	in, err := input(cmd, ks, args)
	if err != nil {
		return err
	}
	out, err := process(ks, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
