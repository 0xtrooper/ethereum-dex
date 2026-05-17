// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package trade

import (
	marketcmd "dex/cmd/trade/market"
	ordercmd "dex/cmd/trade/order"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trade",
		Aliases: []string{"orderbook"},
		Short:   "Interact with the exchange order book",
	}

	cmd.AddCommand(marketcmd.NewCommand(cfg, ks))
	cmd.AddCommand(ordercmd.NewCommand(cfg, ks))

	return cmd
}
