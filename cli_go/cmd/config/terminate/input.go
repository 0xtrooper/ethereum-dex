package configterminate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type dataIn struct {
	Path string
}

func input(cmd *cobra.Command, svc ConfigTerminator) (*dataIn, error) {
	path := svc.Path()

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStdout(), "This will permanently delete the configuration file at:\n  %s\n\nType 'terminate' to confirm: ", path)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.TrimSpace(scanner.Text()) != "terminate" {
			return nil, fmt.Errorf("aborted")
		}
	}

	return &dataIn{Path: path}, nil
}
