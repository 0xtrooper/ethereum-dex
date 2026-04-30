package order

import (
	"context"
	"fmt"
	"math/big"

	"dex/service"

	"github.com/spf13/cobra"
)

type getIn struct {
	contractAddress string
	orderID         *big.Int
}

type getOut struct {
	user          string
	baseToken     string
	quoteToken    string
	side          string
	baseQuantity  string
	quoteQuantity string
}

func newGetCommand(cfg *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, cfg)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("id", "", "Order id")
	return cmd
}

func runGet(cmd *cobra.Command, cfg *service.Service) error {
	in, err := inputGet(cmd, cfg)
	if err != nil {
		return err
	}
	out, err := processGet(in, cfg)
	if err != nil {
		return err
	}
	return outputGet(cmd, out)
}

func inputGet(cmd *cobra.Command, cfg *service.Service) (*getIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	orderID, err := orderReadBigIntFlag(cmd, "id", "Order id", true)
	if err != nil {
		return nil, err
	}
	return &getIn{contractAddress: contractAddress, orderID: orderID}, nil
}

func processGet(in *getIn, cfg *service.Service) (*getOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	orderbookService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	orderValue, err := orderbookService.FetchOrder(context.Background(), in.orderID)
	if err != nil {
		return nil, err
	}

	side := "unknown"
	if orderValue.Side == service.OrderSideBuy {
		side = "buy"
	} else if orderValue.Side == service.OrderSideSell {
		side = "sell"
	}

	return &getOut{
		user:          orderValue.User.Hex(),
		baseToken:     orderValue.BaseToken.Hex(),
		quoteToken:    orderValue.QuoteToken.Hex(),
		side:          side,
		baseQuantity:  orderValue.BaseQuantity.String(),
		quoteQuantity: orderValue.QuoteQuantity.String(),
	}, nil
}

func outputGet(cmd *cobra.Command, out *getOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "user: %s\n", out.user)
	fmt.Fprintf(cmd.OutOrStdout(), "base_token: %s\n", out.baseToken)
	fmt.Fprintf(cmd.OutOrStdout(), "quote_token: %s\n", out.quoteToken)
	fmt.Fprintf(cmd.OutOrStdout(), "side: %s\n", out.side)
	fmt.Fprintf(cmd.OutOrStdout(), "base_quantity: %s\n", out.baseQuantity)
	fmt.Fprintf(cmd.OutOrStdout(), "quote_quantity: %s\n", out.quoteQuantity)
	return nil
}
