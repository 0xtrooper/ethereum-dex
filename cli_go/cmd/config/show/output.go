package show

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v2"
)

func output(cmd *cobra.Command, out *dataOut) error {
	if !out.Exists {
		fmt.Fprintf(cmd.OutOrStdout(), "No config found. Showing defaults.\n\n")
	}
	data, err := yaml.Marshal(out.Config)
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), string(data))
	return nil
}
