package walletexport

import "github.com/spf13/cobra"

type WalletExporter interface {
	List() []string
	Export(address string, password string) (string, error)
}

func NewCommand(ks WalletExporter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [address]",
		Short: "Export a wallet private key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd, ks, args)
		},
	}
	cmd.Flags().String("password", "", "Keystore password")
	return cmd
}

func Run(cmd *cobra.Command, ks WalletExporter, args []string) error {
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
