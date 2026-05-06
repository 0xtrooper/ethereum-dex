package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type ChainToken struct {
	Symbol   string
	Address  string
	Decimals uint8
}

func DefaultTokensForChain(chainID int64) []ChainToken {
	switch chainID {
	case 1:
		return []ChainToken{
			{Symbol: "WETH", Address: "0xC02aaA39b223FE8D0A0E5C4F27eAD9083C756Cc2", Decimals: 18},
			{Symbol: "USDC", Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", Decimals: 6},
			{Symbol: "USDT", Address: "0xdAC17F958D2ee523a2206206994597C13D831ec7", Decimals: 6},
			{Symbol: "DAI", Address: "0x6B175474E89094C44Da98b954EedeAC495271d0F", Decimals: 18},
			{Symbol: "WBTC", Address: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", Decimals: 8},
			{Symbol: "LINK", Address: "0x514910771AF9Ca656af840dff83E8264EcF986CA", Decimals: 18},
			{Symbol: "UNI", Address: "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", Decimals: 18},
			{Symbol: "AAVE", Address: "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DdAe9", Decimals: 18},
		}
	default:
		return nil
	}
}

func MergeKnownAndConfiguredTokens(chainID int64, configured []TokenConfig) []ChainToken {
	seen := make(map[string]bool)
	out := make([]ChainToken, 0, len(configured)+8)

	for _, t := range configured {
		addr := strings.TrimSpace(t.Address)
		if addr == "" {
			continue
		}
		key := strings.ToLower(addr)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ChainToken{
			Symbol:   strings.TrimSpace(t.Symbol),
			Address:  addr,
			Decimals: t.Decimals,
		})
	}

	for _, t := range DefaultTokensForChain(chainID) {
		key := strings.ToLower(t.Address)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, t)
	}

	return out
}

func FindTrackedToken(chainID int64, configured []TokenConfig, address string) (ChainToken, bool) {
	key := strings.ToLower(strings.TrimSpace(address))
	if key == "" {
		return ChainToken{}, false
	}

	for _, token := range MergeKnownAndConfiguredTokens(chainID, configured) {
		if strings.EqualFold(strings.TrimSpace(token.Address), key) {
			return token, true
		}
	}

	return ChainToken{}, false
}

func FormatTokenRef(symbol string, address string) string {
	trimmedSymbol := strings.TrimSpace(symbol)
	trimmedAddress := strings.TrimSpace(address)
	if trimmedSymbol == "" {
		return trimmedAddress
	}
	return fmt.Sprintf("%s [%s]", trimmedSymbol, trimmedAddress)
}

// ResolveTokenMetadata returns token metadata using config/default tokens first.
// If metadata is incomplete, it queries the chain and persists a local cache entry.
func (s *Service) ResolveTokenMetadata(ctx context.Context, rpc *RPC, chainID int64, tokenAddress common.Address) (ChainToken, error) {
	address := tokenAddress.Hex()
	token, found := FindTrackedToken(chainID, s.Get().Tokens, address)
	if !found {
		token = ChainToken{Address: address}
	}

	needsDecimals := token.Decimals == 0
	needsSymbol := strings.TrimSpace(token.Symbol) == ""
	if !needsDecimals && !needsSymbol {
		return token, nil
	}

	tokenService, err := NewTokenService(rpc, nil, tokenAddress, chainID)
	if err != nil {
		return token, err
	}

	var firstErr error

	if needsSymbol {
		symbol, err := tokenService.Symbol(ctx)
		if err == nil && strings.TrimSpace(symbol) != "" {
			token.Symbol = strings.TrimSpace(symbol)
		} else if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if needsDecimals {
		decimals, err := tokenService.Decimals(ctx)
		if err == nil {
			token.Decimals = decimals
		} else if firstErr == nil {
			firstErr = err
		}
	}

	s.cacheTokenMetadata(token)
	if token.Decimals == 0 {
		if firstErr == nil {
			firstErr = fmt.Errorf("failed to resolve token decimals for %s", address)
		}
		return token, firstErr
	}
	return token, nil
}

func (s *Service) cacheTokenMetadata(token ChainToken) {
	if s == nil || s.config == nil {
		return
	}

	address := strings.TrimSpace(token.Address)
	if !common.IsHexAddress(address) {
		return
	}
	address = common.HexToAddress(address).Hex()

	for i := range s.config.Tokens {
		if !strings.EqualFold(strings.TrimSpace(s.config.Tokens[i].Address), address) {
			continue
		}
		if strings.TrimSpace(token.Symbol) != "" {
			s.config.Tokens[i].Symbol = strings.TrimSpace(token.Symbol)
		}
		if token.Decimals > 0 {
			s.config.Tokens[i].Decimals = token.Decimals
		}
		_ = s.Save(s.config)
		return
	}

	s.config.Tokens = append(s.config.Tokens, TokenConfig{
		Symbol:   strings.TrimSpace(token.Symbol),
		Address:  address,
		Decimals: token.Decimals,
	})
	_ = s.Save(s.config)
}
