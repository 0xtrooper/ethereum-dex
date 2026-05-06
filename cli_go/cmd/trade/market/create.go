package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const marketRPCConnectTimeout = 15 * time.Second

type createIn struct {
	contractAddress string
	walletAddress   string
	baseToken       common.Address
	quoteToken      common.Address
}

type createOut struct {
	marketDeployed bool
	bankAddress    string
	txHash         string
}

func newCreateCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a market for a base/quote pair",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("base", "", "Base token address")
	cmd.Flags().String("quote", "", "Quote token address")
	return cmd
}

func runCreate(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputCreate(cmd, cfg, ks)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processCreate(cmd, in, cfg, ks)
	if err != nil {
		return err
	}
	return outputCreate(cmd, out)
}

func inputCreate(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) (*createIn, error) {
	contractAddress, err := marketReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	walletAddress, _ := cmd.Flags().GetString("wallet")
	walletAddress, err = marketSelectWalletAddress(ks, strings.TrimSpace(walletAddress))
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
	return &createIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
	}, nil
}

func processCreate(cmd *cobra.Command, in *createIn, cfg *service.Service, ks *service.Keystore) (*createOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, marketRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()
	orderbookReadService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	bankAddress, deployed, err := orderbookReadService.FetchBank(ctx, in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}
	if deployed {
		return &createOut{
			marketDeployed: true,
			bankAddress:    bankAddress.Hex(),
		}, nil
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Review create market")
	fmt.Fprintf(cmd.OutOrStdout(), "  Wallet: %s\n", in.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "  Base:   %s\n", in.baseToken.Hex())
	fmt.Fprintf(cmd.OutOrStdout(), "  Quote:  %s\n", in.quoteToken.Hex())
	ok, err := prompt.Confirm("Proceed and send create market transaction")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("aborted")
	}
	walletService, err := service.NewWallet(ks, in.walletAddress)
	if err != nil {
		return nil, err
	}
	orderbookService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}
	tx, err := orderbookService.CreateMarket(ctx, in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}
	return &createOut{
		marketDeployed: false,
		bankAddress:    bankAddress.Hex(),
		txHash:         tx.Hash().Hex(),
	}, nil
}

func outputCreate(cmd *cobra.Command, out *createOut) error {
	if out.marketDeployed {
		fmt.Fprintf(cmd.OutOrStdout(), "Market is already deployed. Bank: %s\n", out.bankAddress)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Create market tx: %s\n", out.txHash)
	return nil
}
