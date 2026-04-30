package order

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

type cancelIn struct {
	contractAddress string
	walletAddress   string
	orderID         *big.Int
}

type cancelOut struct {
	walletAddress string
	txHash        string
}

func newCancelCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an order you own",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("id", "", "Order id")
	return cmd
}

func runCancel(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputCancel(cmd, cfg)
	if err != nil {
		return err
	}
	out, err := processCancel(in, cfg, ks)
	if err != nil {
		return err
	}
	return outputCancel(cmd, out)
}

func inputCancel(cmd *cobra.Command, cfg *service.Service) (*cancelIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	walletAddress, _ := cmd.Flags().GetString("wallet")
	walletAddress = strings.TrimSpace(walletAddress)
	orderID, err := orderReadBigIntFlag(cmd, "id", "Order id", true)
	if err != nil {
		return nil, err
	}
	return &cancelIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		orderID:         orderID,
	}, nil
}

func processCancel(in *cancelIn, cfg *service.Service, ks *service.Keystore) (*cancelOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	walletService, err := service.NewWallet(ks, in.walletAddress)
	if err != nil {
		return nil, err
	}
	orderbookService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	tx, err := orderbookService.CancelOrder(context.Background(), in.orderID)
	if err != nil {
		return nil, err
	}
	return &cancelOut{
		walletAddress: walletService.Address(),
		txHash:        tx.Hash().Hex(),
	}, nil
}

func outputCancel(cmd *cobra.Command, out *cancelOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", out.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel tx: %s\n", out.txHash)
	return nil
}
