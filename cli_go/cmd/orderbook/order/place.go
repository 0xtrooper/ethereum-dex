package order

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const orderRPCConnectTimeout = 15 * time.Second

type placeIn struct {
	contractAddress string
	walletAddress   string
	baseToken       common.Address
	quoteToken      common.Address
	side            uint8
	baseQuantity    *big.Int
	quoteQuantity   *big.Int
	value           *big.Int
}

type placeOut struct {
	signerWallet   string
	bankAddress    string
	createdMarket  bool
	createMarketTx string
	placeOrderTx   string
}

func newPlaceCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place",
		Short: "Place a new order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlace(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("base", "", "Base token address")
	cmd.Flags().String("quote", "", "Quote token address")
	cmd.Flags().String("side", "", "Order side: buy or sell")
	cmd.Flags().String("base-qty", "", "Base token quantity (integer, wei-style units)")
	cmd.Flags().String("quote-qty", "", "Quote token quantity (integer, wei-style units)")
	cmd.Flags().String("value-wei", "", "Native value in wei for payable path (optional)")
	return cmd
}

func runPlace(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputPlace(cmd, cfg)
	if err != nil {
		return err
	}
	out, err := processPlace(in, cfg, ks)
	if err != nil {
		return err
	}
	return outputPlace(cmd, out)
}

func inputPlace(cmd *cobra.Command, cfg *service.Service) (*placeIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	walletAddress, _ := cmd.Flags().GetString("wallet")
	walletAddress = strings.TrimSpace(walletAddress)
	baseToken, err := orderReadAddressFlag(cmd, "base", "Base token address")
	if err != nil {
		return nil, err
	}
	quoteToken, err := orderReadAddressFlag(cmd, "quote", "Quote token address")
	if err != nil {
		return nil, err
	}
	side, err := orderReadSide(cmd)
	if err != nil {
		return nil, err
	}
	baseQuantity, err := orderReadBigIntFlag(cmd, "base-qty", "Base quantity", true)
	if err != nil {
		return nil, err
	}
	quoteQuantity, err := orderReadBigIntFlag(cmd, "quote-qty", "Quote quantity", true)
	if err != nil {
		return nil, err
	}
	value, err := orderReadBigIntFlag(cmd, "value-wei", "Native value in wei", false)
	if err != nil {
		return nil, err
	}

	return &placeIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
		side:            side,
		baseQuantity:    baseQuantity,
		quoteQuantity:   quoteQuantity,
		value:           value,
	}, nil
}

func processPlace(in *placeIn, cfg *service.Service, ks *service.Keystore) (*placeOut, error) {
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

	ctx := context.Background()
	bankAddress, deployed, err := orderbookService.FetchBank(ctx, in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}

	out := &placeOut{
		signerWallet: walletService.Address(),
		bankAddress:  bankAddress.Hex(),
	}

	if !deployed {
		ok, err := prompt.Confirm("Market is not deployed for this pair. Deploy market now")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("aborted")
		}

		createTx, err := orderbookService.CreateMarket(ctx, in.baseToken, in.quoteToken)
		if err != nil {
			return nil, err
		}
		out.createdMarket = true
		out.createMarketTx = createTx.Hash().Hex()
	}

	placeTx, err := orderbookService.PlaceOrder(ctx, in.baseToken, in.quoteToken, in.side, in.baseQuantity, in.quoteQuantity, in.value)
	if err != nil {
		return nil, err
	}
	out.placeOrderTx = placeTx.Hash().Hex()

	return out, nil
}

func outputPlace(cmd *cobra.Command, out *placeOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", out.signerWallet)
	fmt.Fprintf(cmd.OutOrStdout(), "Market bank: %s\n", out.bankAddress)
	if out.createdMarket {
		fmt.Fprintf(cmd.OutOrStdout(), "Created market tx: %s\n", out.createMarketTx)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Place order tx: %s\n", out.placeOrderTx)
	return nil
}
