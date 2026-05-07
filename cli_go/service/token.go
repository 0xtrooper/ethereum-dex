package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/contracts/erc20"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const erc20ReadABI = `[
  {
    "inputs": [
      { "internalType": "address", "name": "owner", "type": "address" },
      { "internalType": "address", "name": "spender", "type": "address" }
    ],
    "name": "allowance",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "symbol",
    "outputs": [{ "internalType": "string", "name": "", "type": "string" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "decimals",
    "outputs": [{ "internalType": "uint8", "name": "", "type": "uint8" }],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{ "internalType": "address", "name": "account", "type": "address" }],
    "name": "balanceOf",
    "outputs": [{ "internalType": "uint256", "name": "", "type": "uint256" }],
    "stateMutability": "view",
    "type": "function"
  }
]`

type TokenService struct {
	address  common.Address
	chainID  *big.Int
	rpc      *RPC
	wallet   *Wallet
	contract *erc20.ERC20
}

func NewTokenService(rpc *RPC, wallet *Wallet, tokenAddress common.Address, chainID int64) (*TokenService, error) {
	if rpc == nil || rpc.Client() == nil {
		return nil, fmt.Errorf("rpc service is not initialized")
	}
	if chainID <= 0 {
		return nil, fmt.Errorf("chain id must be greater than zero")
	}

	contract, err := erc20.NewERC20(tokenAddress, rpc.Client())
	if err != nil {
		return nil, fmt.Errorf("failed to bind token contract: %w", err)
	}

	return &TokenService{
		address:  tokenAddress,
		chainID:  big.NewInt(chainID),
		rpc:      rpc,
		wallet:   wallet,
		contract: contract,
	}, nil
}

func (s *TokenService) Address() string {
	if s == nil {
		return ""
	}
	return s.address.Hex()
}

func (s *TokenService) Approve(ctx context.Context, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("token service is not initialized")
	}
	if err := requirePositive(amount, "amount"); err != nil {
		return nil, err
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return nil, err
	}

	return s.contract.Approve(opts, spender, amount)
}

func (s *TokenService) Send(ctx context.Context, to common.Address, amount *big.Int) (*types.Transaction, error) {
	if s == nil || s.contract == nil {
		return nil, fmt.Errorf("token service is not initialized")
	}
	if err := requirePositive(amount, "amount"); err != nil {
		return nil, err
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return nil, err
	}

	return s.contract.Transfer(opts, to, amount)
}

func (s *TokenService) Allowance(ctx context.Context, owner common.Address, spender common.Address) (*big.Int, error) {
	if s == nil || s.rpc == nil || s.rpc.Client() == nil {
		return nil, fmt.Errorf("token service is not initialized")
	}
	return s.callUint256View(ctx, "allowance", owner, spender)
}

func (s *TokenService) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	if s == nil || s.rpc == nil || s.rpc.Client() == nil {
		return nil, fmt.Errorf("token service is not initialized")
	}
	return s.callUint256View(ctx, "balanceOf", account)
}

func (s *TokenService) Symbol(ctx context.Context) (string, error) {
	if s == nil || s.rpc == nil || s.rpc.Client() == nil {
		return "", fmt.Errorf("token service is not initialized")
	}
	return s.callStringView(ctx, "symbol")
}

func (s *TokenService) Decimals(ctx context.Context) (uint8, error) {
	if s == nil || s.rpc == nil || s.rpc.Client() == nil {
		return 0, fmt.Errorf("token service is not initialized")
	}
	return s.callUint8View(ctx, "decimals")
}

func (s *TokenService) transactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(s.wallet.PrivateKey(), s.chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	return opts, nil
}

func (s *TokenService) callUint256View(ctx context.Context, method string, args ...interface{}) (*big.Int, error) {
	values, err := s.callView(ctx, method, args...)
	if err != nil {
		return nil, err
	}
	value, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected %s response type %T", method, values[0])
	}
	return value, nil
}

func (s *TokenService) callStringView(ctx context.Context, method string, args ...interface{}) (string, error) {
	values, err := s.callView(ctx, method, args...)
	if err != nil {
		return "", err
	}
	value, ok := values[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected %s response type %T", method, values[0])
	}
	return value, nil
}

func (s *TokenService) callUint8View(ctx context.Context, method string, args ...interface{}) (uint8, error) {
	values, err := s.callView(ctx, method, args...)
	if err != nil {
		return 0, err
	}
	value, ok := values[0].(uint8)
	if !ok {
		return 0, fmt.Errorf("unexpected %s response type %T", method, values[0])
	}
	return value, nil
}

func (s *TokenService) callView(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ReadABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse erc20 read abi: %w", err)
	}

	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s call: %w", method, err)
	}

	callMsg := ethereum.CallMsg{
		To:   &s.address,
		Data: data,
	}
	result, err := s.rpc.Client().CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("%s call failed: %w", method, err)
	}

	values, err := parsedABI.Unpack(method, result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s response: %w", method, err)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected %s response length: %d", method, len(values))
	}
	return values, nil
}
