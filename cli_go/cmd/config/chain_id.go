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
	"strconv"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

func newChainIDCommand(svc *service.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "chain-id <id>",
		Short: "Set the Ethereum chain ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawChainID := strings.TrimSpace(args[0])
			chainID, err := strconv.ParseInt(rawChainID, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain id %q: %w", rawChainID, err)
			}
			if chainID <= 0 {
				return fmt.Errorf("chain id must be greater than zero")
			}

			cfg, created, err := svc.Ensure()
			if err != nil {
				return err
			}
			cfg.Network.ChainID = chainID
			if err := svc.Save(cfg); err != nil {
				return err
			}

			if created {
				fmt.Fprintf(cmd.OutOrStdout(), "Created config: %s\n", svc.Path())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set chain_id: %d\n", cfg.Network.ChainID)
			return nil
		},
	}
}
