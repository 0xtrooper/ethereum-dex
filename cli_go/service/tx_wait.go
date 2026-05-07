package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const DefaultTxWaitTimeout = 90 * time.Second

// WaitForTxReceipt polls for a transaction receipt until it is available or the timeout/context expires.
func WaitForTxReceipt(parent context.Context, rpc *RPC, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	if rpc == nil || rpc.Client() == nil {
		return nil, fmt.Errorf("rpc service is not initialized")
	}

	if timeout <= 0 {
		timeout = DefaultTxWaitTimeout
	}

	waitCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		receipt, err := rpc.Client().TransactionReceipt(waitCtx, txHash)
		if err == nil {
			return receipt, nil
		}
		if !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}

		select {
		case <-waitCtx.Done():
			return nil, fmt.Errorf("timed out waiting for transaction %s to be mined", txHash.Hex())
		case <-ticker.C:
		}
	}
}

func WaitForTxSuccess(parent context.Context, rpc *RPC, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	receipt, err := WaitForTxReceipt(parent, rpc, txHash, timeout)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, fmt.Errorf("no receipt returned for transaction %s", txHash.Hex())
	}
	if receipt.Status == types.ReceiptStatusFailed {
		return nil, fmt.Errorf("transaction %s reverted in block %d", txHash.Hex(), receipt.BlockNumber.Uint64())
	}
	return receipt, nil
}
