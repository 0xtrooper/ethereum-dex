package walletlist

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	if len(out.Addresses) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No wallets in keystore.")
		return nil
	}
	for i, addr := range out.Addresses {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
	}
	return nil
}
