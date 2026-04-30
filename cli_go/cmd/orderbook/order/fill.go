package order

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/service"

	"github.com/spf13/cobra"
)

type fillIn struct {
	contractAddress string
	walletAddress   string
	orderID         *big.Int
	baseQuantity    *big.Int
	value           *big.Int
}

type fillOut struct {
	walletAddress string
	txHash        string
}

func newFillCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "Fill an existing order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFill(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("id", "", "Order id")
	cmd.Flags().String("base-qty", "", "Fill base quantity (integer, wei-style units)")
	cmd.Flags().String("value-wei", "", "Native value in wei for payable path (optional)")
	return cmd
}

func runFill(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputFill(cmd, cfg)
	if err != nil {
		return err
	}
	out, err := processFill(in, cfg, ks)
	if err != nil {
		return err
	}
	return outputFill(cmd, out)
}

func inputFill(cmd *cobra.Command, cfg *service.Service) (*fillIn, error) {
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
	baseQuantity, err := orderReadBigIntFlag(cmd, "base-qty", "Fill base quantity", true)
	if err != nil {
		return nil, err
	}
	value, err := orderReadBigIntFlag(cmd, "value-wei", "Native value in wei", false)
	if err != nil {
		return nil, err
	}
	return &fillIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		orderID:         orderID,
		baseQuantity:    baseQuantity,
		value:           value,
	}, nil
}

func processFill(in *fillIn, cfg *service.Service, ks *service.Keystore) (*fillOut, error) {
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

	tx, err := orderbookService.FillOrder(context.Background(), in.orderID, in.baseQuantity, in.value)
	if err != nil {
		return nil, err
	}
	return &fillOut{
		walletAddress: walletService.Address(),
		txHash:        tx.Hash().Hex(),
	}, nil
}

func outputFill(cmd *cobra.Command, out *fillOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", out.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "Fill tx: %s\n", out.txHash)
	return nil
}
