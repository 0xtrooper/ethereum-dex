package order

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/internal/amount"
	"dex/service"

	"github.com/spf13/cobra"
)

type getIn struct {
	contractAddress string
	orderID         *big.Int
}

type getOut struct {
	user                 string
	baseToken            string
	quoteToken           string
	baseSymbol           string
	quoteSymbol          string
	side                 string
	baseQuantityRaw      string
	baseQuantityDisplay  string
	quoteQuantityRaw     string
	quoteQuantityDisplay string
}

func newGetCommand(cfg *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a specific order",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				currentID, _ := cmd.Flags().GetString("id")
				if strings.TrimSpace(currentID) == "" {
					if err := cmd.Flags().Set("id", strings.TrimSpace(args[0])); err != nil {
						return err
					}
				}
			}
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
	cmd.SilenceUsage = true
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

	ctx := context.Background()
	orderValue, err := orderbookService.FetchOrder(ctx, in.orderID)
	if err != nil {
		return nil, err
	}

	side := "unknown"
	if orderValue.Side == service.OrderSideBuy {
		side = "buy"
	} else if orderValue.Side == service.OrderSideSell {
		side = "sell"
	}

	baseDecimals := uint8(18)
	baseSymbol := ""
	if meta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, orderValue.BaseToken); err == nil {
		if meta.Decimals > 0 {
			baseDecimals = meta.Decimals
		}
		baseSymbol = strings.TrimSpace(meta.Symbol)
	}
	quoteDecimals := uint8(18)
	quoteSymbol := ""
	if meta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, orderValue.QuoteToken); err == nil {
		if meta.Decimals > 0 {
			quoteDecimals = meta.Decimals
		}
		quoteSymbol = strings.TrimSpace(meta.Symbol)
	}

	return &getOut{
		user:                 orderValue.User.Hex(),
		baseToken:            service.FormatTokenRef(baseSymbol, orderValue.BaseToken.Hex()),
		quoteToken:           service.FormatTokenRef(quoteSymbol, orderValue.QuoteToken.Hex()),
		baseSymbol:           baseSymbol,
		quoteSymbol:          quoteSymbol,
		side:                 side,
		baseQuantityRaw:      orderValue.BaseQuantity.String(),
		baseQuantityDisplay:  amount.FormatUnits(orderValue.BaseQuantity, baseDecimals),
		quoteQuantityRaw:     orderValue.QuoteQuantity.String(),
		quoteQuantityDisplay: amount.FormatUnits(orderValue.QuoteQuantity, quoteDecimals),
	}, nil
}

func outputGet(cmd *cobra.Command, out *getOut) error {
	sideDetail := ""
	if out.side == "sell" && out.quoteSymbol != "" && out.baseSymbol != "" {
		sideDetail = fmt.Sprintf(" (%s -> %s)", out.quoteSymbol, out.baseSymbol)
	} else if out.side == "buy" && out.baseSymbol != "" && out.quoteSymbol != "" {
		sideDetail = fmt.Sprintf(" (%s -> %s)", out.baseSymbol, out.quoteSymbol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Maker: %s\n", out.user)
	fmt.Fprintf(cmd.OutOrStdout(), "Base Token: %s\n", out.baseToken)
	fmt.Fprintf(cmd.OutOrStdout(), "Quote Token: %s\n", out.quoteToken)
	fmt.Fprintf(cmd.OutOrStdout(), "Side: %s%s\n", out.side, sideDetail)
	fmt.Fprintf(cmd.OutOrStdout(), "Base Quantity: %s (raw: %s)\n", out.baseQuantityDisplay, out.baseQuantityRaw)
	fmt.Fprintf(cmd.OutOrStdout(), "Quote Quantity: %s (raw: %s)\n", out.quoteQuantityDisplay, out.quoteQuantityRaw)
	return nil
}
