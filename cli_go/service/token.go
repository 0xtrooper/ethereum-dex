package service

import (
	"context"
	"fmt"
	"math/big"

	"dex/contracts/erc20"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TokenService struct {
	address  common.Address
	chainID  *big.Int
	wallet   *Wallet
	contract *erc20.ERC20
}

func NewTokenService(rpc *RPC, wallet *Wallet, tokenAddress common.Address, chainID int64) (*TokenService, error) {
	if rpc == nil || rpc.Client() == nil {
		return nil, fmt.Errorf("rpc service is not initialized")
	}
	if wallet == nil || wallet.PrivateKey() == nil {
		return nil, fmt.Errorf("wallet service is not initialized")
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

func (s *TokenService) transactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(s.wallet.PrivateKey(), s.chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	return opts, nil
}
