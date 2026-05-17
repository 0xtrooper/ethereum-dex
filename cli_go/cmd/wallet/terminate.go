// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
