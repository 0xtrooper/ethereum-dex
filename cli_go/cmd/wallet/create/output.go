package walletcreate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Created wallet: %s\n", out.Address)
	return nil
}
