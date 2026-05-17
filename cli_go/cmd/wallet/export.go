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
	"strconv"
	"strings"

	"dex/internal/prompt"
	"dex/service"

	"github.com/spf13/cobra"
)

const walletExportWarning = `
WARNING - SENSITIVE DATA WILL BE DISPLAYED

Your unencrypted private key is about to be printed to your terminal in plain text. Before proceeding, be aware that:

- Anyone who can view your screen, access your terminal history, read your shell logs, or inspect your clipboard may be able to read the key.
- Exposing your private key may result in the immediate and permanent loss of any funds held at the associated address.
- This software accepts no liability for any loss of funds or data resulting from the exposure or misuse of an exported private key.
`

func newWalletExportCommand(ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [address]",
		Short: "Export a wallet private key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(os.Stderr, walletExportWarning)
			agreed, err := prompt.Confirm("I understand the risks and wish to proceed")
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return err
			}
			if !agreed {
				return fmt.Errorf("aborted")
			}

			wallets := ks.List()
			var address string
			if len(args) > 0 {
				address = strings.TrimSpace(args[0])
			} else {
				address, err = promptWalletExportSelection(cmd, wallets)
				if err != nil {
					return err
				}
			}

			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				password, err = prompt.Password("Password to decrypt your keystore")
				if err != nil {
					return err
				}
			}

			privateKey, err := ks.Export(address, password)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Private Key:", privateKey)
			return nil
		},
	}

	cmd.Flags().String("password", "", "Keystore password")
	return cmd
}

func promptWalletExportSelection(cmd *cobra.Command, wallets []string) (string, error) {
	if len(wallets) == 0 {
		return "", fmt.Errorf("no wallets in keystore")
	}
	if len(wallets) == 1 {
		return wallets[0], nil
	}

	for i, addr := range wallets {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "Select wallet to export: ")
		scanner.Scan()
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(wallets) {
			fmt.Fprintf(os.Stderr, "Please enter a number between 1 and %d.\n", len(wallets))
			continue
		}
		return wallets[n-1], nil
	}
}
