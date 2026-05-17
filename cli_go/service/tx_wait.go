// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
