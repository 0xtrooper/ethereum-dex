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
	"regexp"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

var ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

func newContractCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "contract <address>",
		Short: "Set the DEX contract address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if !ethAddressPattern.MatchString(address) {
				return fmt.Errorf("invalid Ethereum address %q", address)
			}

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}
			cfg.Contract.Address = address
			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set contract address: %s\n", cfg.Contract.Address)
			return nil
		},
	}
}
