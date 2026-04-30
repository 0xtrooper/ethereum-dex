package orderbook

import (
	marketcmd "dex/cmd/orderbook/market"
	ordercmd "dex/cmd/orderbook/order"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderbook",
		Short: "Interact with the exchange order book",
	}

	cmd.AddCommand(marketcmd.NewCommand(cfg, ks))
	cmd.AddCommand(ordercmd.NewCommand(cfg, ks))

	return cmd
}
