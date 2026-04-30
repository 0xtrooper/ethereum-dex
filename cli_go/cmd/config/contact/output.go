package contact

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	if out.Created {
		fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", out.ConfigPath)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Set contract address: %s\n", out.Address)
	return nil
}
