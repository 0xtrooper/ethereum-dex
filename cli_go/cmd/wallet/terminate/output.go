package walletterminate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d wallet(s).\n", out.Count)
	return nil
}
