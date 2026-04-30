package walletexport

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	fmt.Fprintln(cmd.OutOrStdout(), out.PrivateKey)
	return nil
}
