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

	"dex/service"

	"github.com/spf13/cobra"
)

func newWalletDeleteCommand(ks *service.Keystore) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [address...]",
		Short: "Delete one or more wallets from the keystore",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			addresses := args
			if len(addresses) == 0 {
				wallets := ks.List()
				if len(wallets) == 0 {
					return fmt.Errorf("no wallets in keystore")
				}

				for i, addr := range wallets {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
				}

				selected, err := promptWalletDeleteSelection(wallets)
				if err != nil {
					return err
				}
				addresses = selected
			}

			var firstErr error
			for _, addr := range addresses {
				err := ks.Delete(addr)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to delete %s: %v\n", addr, err)
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Deleted wallet: %s\n", addr)
			}
			return firstErr
		},
	}
}

func promptWalletDeleteSelection(wallets []string) ([]string, error) {
	fmt.Fprint(os.Stderr, "Select wallets to delete (e.g. 1,3): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	raw := scanner.Text()

	var selected []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", part)
		}
		if n < 1 || n > len(wallets) {
			return nil, fmt.Errorf("selection %d out of range (1-[%d]", n, len(wallets))
		}
		selected = append(selected, wallets[n-1])
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no wallets selected")
	}
	return selected, nil
}
