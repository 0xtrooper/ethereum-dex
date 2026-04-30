package walletdelete

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type dataIn struct {
	Addresses []string
}

func input(cmd *cobra.Command, args []string, wallets []string) (*dataIn, error) {
	if len(args) > 0 {
		return &dataIn{Addresses: args}, nil
	}

	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets in keystore")
	}

	for i, addr := range wallets {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
	}

	selected, err := promptSelection(wallets)
	if err != nil {
		return nil, err
	}

	return &dataIn{Addresses: selected}, nil
}

func promptSelection(wallets []string) ([]string, error) {
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
