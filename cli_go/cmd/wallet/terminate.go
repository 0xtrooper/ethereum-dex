package wallet

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newWalletTerminateCommand(ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Delete all wallets from the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wallets := ks.List()
			if len(wallets) == 0 {
				return fmt.Errorf("no wallets in keystore")
			}

			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(), "This will permanently delete %d wallet(s). Type 'terminate' to confirm: ", len(wallets))
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if strings.TrimSpace(scanner.Text()) != "terminate" {
					return fmt.Errorf("aborted")
				}
			}

			if err := ks.DeleteAll(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d wallet(s).\n", len(wallets))
			return nil
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}
