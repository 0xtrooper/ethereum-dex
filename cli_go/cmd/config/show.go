// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package config

import (
	"fmt"

	"dex/service"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v2"
)

func newShowCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show dex configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, exists, err := svc.LoadIfExists()
			if err != nil {
				return err
			}
			if !exists {
				fmt.Fprintf(cmd.OutOrStdout(), "No config found. Showing defaults.\n\n")
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}
