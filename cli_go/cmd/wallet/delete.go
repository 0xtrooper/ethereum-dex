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
