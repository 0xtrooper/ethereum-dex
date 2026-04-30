package walletdelete

import (
	"fmt"

	"github.com/spf13/cobra"
)

func output(cmd *cobra.Command, out *dataOut) error {
	var firstErr error
	for _, r := range out.Results {
		if r.Err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Failed to delete %s: %v\n", r.Address, r.Err)
			if firstErr == nil {
				firstErr = r.Err
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted wallet: %s\n", r.Address)
		}
	}
	return firstErr
}
