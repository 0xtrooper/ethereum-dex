package order

import (
	"context"
	"fmt"

	"dex/service"

	"github.com/spf13/cobra"
)

type countIn struct {
	contractAddress string
}

type countOut struct {
	count string
}

func newCountCommand(cfg *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get total order counter",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(cmd, cfg)
		},
	}
	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	return cmd
}

func runCount(cmd *cobra.Command, cfg *service.Service) error {
	in, err := inputCount(cmd, cfg)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processCount(in, cfg)
	if err != nil {
		return err
	}
	return outputCount(cmd, out)
}

func inputCount(cmd *cobra.Command, cfg *service.Service) (*countIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	return &countIn{contractAddress: contractAddress}, nil
}

func processCount(in *countIn, cfg *service.Service) (*countOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	orderbookService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	count, err := orderbookService.FetchOrderCounter(context.Background())
	if err != nil {
		return nil, err
	}
	return &countOut{count: count.String()}, nil
}

func outputCount(cmd *cobra.Command, out *countOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "order count: %s\n", out.count)
	return nil
}
