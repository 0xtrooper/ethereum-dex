// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package market

import (
	"context"
	"fmt"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

type getIn struct {
	contractAddress string
	baseToken       common.Address
	quoteToken      common.Address
}

type getOut struct {
	deployed bool
	bank     string
}

func newGetCommand(cfg *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get market deployment status for a base/quote pair",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, cfg)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("base", "", "Base token address")
	cmd.Flags().String("quote", "", "Quote token address")
	return cmd
}

func runGet(cmd *cobra.Command, cfg *service.Service) error {
	in, err := inputGet(cmd, cfg)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processGet(in, cfg)
	if err != nil {
		return err
	}
	return outputGet(cmd, out)
}

func inputGet(cmd *cobra.Command, cfg *service.Service) (*getIn, error) {
	contractAddress, err := marketReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	baseToken, err := marketReadAddressFlag(cmd, "base", "Base token address")
	if err != nil {
		return nil, err
	}
	quoteToken, err := marketReadAddressFlag(cmd, "quote", "Quote token address")
	if err != nil {
		return nil, err
	}
	return &getIn{
		contractAddress: contractAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
	}, nil
}

func processGet(in *getIn, cfg *service.Service) (*getOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, marketRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()
	orderbookService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}
	bankAddress, deployed, err := orderbookService.FetchBank(context.Background(), in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}
	return &getOut{
		deployed: deployed,
		bank:     bankAddress.Hex(),
	}, nil
}

func outputGet(cmd *cobra.Command, out *getOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "deployed: %t\n", out.deployed)
	fmt.Fprintf(cmd.OutOrStdout(), "bank: %s\n", out.bank)
	return nil
}
