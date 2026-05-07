package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newTerminateCommand(svc *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Delete the dex configuration file from disk",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := svc.Path()
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(), "This will permanently delete the configuration file at:\n  %s\n\nType 'terminate' to confirm: ", path)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if strings.TrimSpace(scanner.Text()) != "terminate" {
					return fmt.Errorf("aborted")
				}
			}

			if err := svc.Delete(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted configuration: %s\n", path)
			return nil
		},
	}
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}
