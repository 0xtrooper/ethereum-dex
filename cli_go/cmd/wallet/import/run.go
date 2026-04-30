package walletimport

import "github.com/spf13/cobra"

type WalletImporter interface {
	Import(privateKeyHex string, password string) (string, error)
	CreateFromMnemonic(mnemonic string, password string) (string, error)
}

func NewCommand(ks WalletImporter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a wallet into the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
	cmd.Flags().String("password", "", "Keystore password")
	cmd.Flags().String("private-key", "", "Private key hex (skips interactive prompt)")
	return cmd
}

func Run(cmd *cobra.Command, ks WalletImporter, args []string) error {
	in, err := input(cmd, args)
	if err != nil {
		return err
	}
	out, err := process(ks, in)
	if err != nil {
		return err
	}
	return output(cmd, out)
}
