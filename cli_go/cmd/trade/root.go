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
