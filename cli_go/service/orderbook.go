package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/contracts/orderbook"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	OrderSideBuy  uint8 = 0
	OrderSideSell uint8 = 1
)

type Order struct {
	User          common.Address
	BaseQuantity  *big.Int
	QuoteQuantity *big.Int
	BaseToken     common.Address
	QuoteToken    common.Address
	Side          uint8
}

type OrderBookService struct {
	address  common.Address
	chainID  *big.Int
	rpc      *RPC
	wallet   *Wallet
	contract *orderbook.OrderBook
}

func NewOrderBookService(rpc *RPC, wallet *Wallet, contractAddress string, chainID int64) (*OrderBookService, error) {
	if rpc == nil || rpc.Client() == nil {
		return nil, fmt.Errorf("rpc service is not initialized")
	}
	if chainID <= 0 {
		return nil, fmt.Errorf("chain id must be greater than zero")
	}

	address, err := parseHexAddress(contractAddress, "contract address")
	if err != nil {
		return nil, err
	}

	contract, err := orderbook.NewOrderBook(address, rpc.Client())
	if err != nil {
		return nil, fmt.Errorf("failed to bind order book contract: %w", err)
	}

	return &OrderBookService{
		address:  address,
		chainID:  big.NewInt(chainID),
		rpc:      rpc,
		wallet:   wallet,
		contract: contract,
	}, nil
}

func (s *OrderBookService) Address() string {
	if s == nil {
		return ""
	}
	return s.address.Hex()
}

func (s *OrderBookService) FetchBank(ctx context.Context, baseToken common.Address, quoteToken common.Address) (common.Address, bool, error) {
	if s == nil || s.contract == nil {
		return common.Address{}, false, fmt.Errorf("order book service is not initialized")
	}

	bankHash := marketBankHash(baseToken, quoteToken)
	bankAddress, err := s.contract.Banks(&bind.CallOpts{Context: ctx}, bankHash)
	if err != nil {
		return common.Address{}, false, err
	}

	deployed := bankAddress != (common.Address{})
	return bankAddress, deployed, nil
}

func (s *OrderBookService) FetchOrderCounter(ctx context.Context) (*big.Int, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}
	return s.contract.OrderCounter(&bind.CallOpts{Context: ctx})
}

func (s *OrderBookService) FetchOrder(ctx context.Context, orderID *big.Int) (*Order, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}
	if err := requirePositive(orderID, "order id"); err != nil {
		return nil, err
	}

	order, err := s.contract.Orders(&bind.CallOpts{Context: ctx}, orderID)
	if err != nil {
		return nil, err
	}

	return &Order{
		User:          order.User,
		BaseQuantity:  order.BaseQuantity,
		QuoteQuantity: order.QuoteQuantity,
		BaseToken:     order.BaseToken,
		QuoteToken:    order.QuoteToken,
		Side:          order.Side,
	}, nil
}

func (s *OrderBookService) CreateMarket(ctx context.Context, baseToken common.Address, quoteToken common.Address) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}

	opts, err := s.transactOpts(ctx, nil)
	if err != nil {
		return nil, err
	}

	return s.contract.CreateMarket(opts, baseToken, quoteToken)
}

func (s *OrderBookService) PlaceOrder(ctx context.Context, baseToken common.Address, quoteToken common.Address, side uint8, baseQuantity *big.Int, quoteQuantity *big.Int, value *big.Int) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}
	if side != OrderSideBuy && side != OrderSideSell {
		return nil, fmt.Errorf("invalid order side %d", side)
	}
	if err := requirePositive(baseQuantity, "base quantity"); err != nil {
		return nil, err
	}
	if err := requirePositive(quoteQuantity, "quote quantity"); err != nil {
		return nil, err
	}

	paymentValue, err := requireNonNegativeOrZero(value, "payment value")
	if err != nil {
		return nil, err
	}

	opts, err := s.transactOpts(ctx, paymentValue)
	if err != nil {
		return nil, err
	}

	return s.contract.PlaceOrder(opts, baseToken, quoteToken, side, baseQuantity, quoteQuantity)
}

func (s *OrderBookService) CancelOrder(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}
	if err := requirePositive(orderID, "order id"); err != nil {
		return nil, err
	}

	opts, err := s.transactOpts(ctx, nil)
	if err != nil {
		return nil, err
	}

	return s.contract.CancelOrder(opts, orderID)
}

func (s *OrderBookService) FillOrder(ctx context.Context, orderID *big.Int, baseQuantity *big.Int, value *big.Int) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("order book service is not initialized")
	}
	if err := requirePositive(orderID, "order id"); err != nil {
		return nil, err
	}
	if err := requirePositive(baseQuantity, "base quantity"); err != nil {
		return nil, err
	}

	paymentValue, err := requireNonNegativeOrZero(value, "payment value")
	if err != nil {
		return nil, err
	}

	opts, err := s.transactOpts(ctx, paymentValue)
	if err != nil {
		return nil, err
	}

	return s.contract.FillOrder(opts, orderID, baseQuantity)
}

func (s *OrderBookService) transactOpts(ctx context.Context, value *big.Int) (*bind.TransactOpts, error) {
	if s.wallet == nil || s.wallet.PrivateKey() == nil {
		return nil, fmt.Errorf("wallet service is not initialized")
	}

	opts, err := bind.NewKeyedTransactorWithChainID(s.wallet.PrivateKey(), s.chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	opts.Value = value
	return opts, nil
}

func parseHexAddress(value string, fieldName string) (common.Address, error) {
	trimmed := strings.TrimSpace(value)
	if !common.IsHexAddress(trimmed) {
		return common.Address{}, fmt.Errorf("invalid %s %q", fieldName, value)
	}
	return common.HexToAddress(trimmed), nil
}

func requirePositive(value *big.Int, fieldName string) error {
	if value == nil {
		return fmt.Errorf("%s is required", fieldName)
	}
	if value.Sign() <= 0 {
		return fmt.Errorf("%s must be greater than zero", fieldName)
	}
	return nil
}

func requireNonNegativeOrZero(value *big.Int, fieldName string) (*big.Int, error) {
	if value == nil {
		return big.NewInt(0), nil
	}
	if value.Sign() < 0 {
		return nil, fmt.Errorf("%s cannot be negative", fieldName)
	}
	return value, nil
}

func marketBankHash(baseToken common.Address, quoteToken common.Address) [32]byte {
	hash := crypto.Keccak256Hash(baseToken.Bytes(), quoteToken.Bytes())
	return [32]byte(hash)
}
