package walletterminate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type dataIn struct {
	Count int
}

func input(cmd *cobra.Command, wallets []string) (*dataIn, error) {
	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets in keystore")
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStdout(), "This will permanently delete %d wallet(s). Type 'terminate' to confirm: ", len(wallets))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "terminate" {
			return nil, fmt.Errorf("aborted")
		}
	}

	return &dataIn{Count: len(wallets)}, nil
}
