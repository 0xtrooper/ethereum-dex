package configterminate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted configuration: %s\n", out.Path)
	return nil
}
