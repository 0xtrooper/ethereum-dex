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
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type RPC struct {
	url    string
	client *ethclient.Client
}

func NewRPC(networkConfig NetworkConfig, timeout time.Duration) (*RPC, error) {
	rpcURL := strings.TrimSpace(networkConfig.RPCURL)
	if rpcURL == "" {
		return nil, fmt.Errorf("rpc url is not configured; set it with `dex config rpc <url>`")
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("rpc timeout must be greater than zero")
	}

	endpointLabel := safeRPCURLLabel(rpcURL)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), timeout)
	client, err := ethclient.DialContext(dialCtx, rpcURL)
	dialCancel()
	if err != nil {
		return nil, fmt.Errorf("failed to dial rpc %q: %w", endpointLabel, err)
	}

	chainCtx, chainCancel := context.WithTimeout(context.Background(), timeout)
	networkChainID, err := client.ChainID(chainCtx)
	chainCancel()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("rpc %q is not reachable: %w", endpointLabel, err)
	}

	if networkConfig.ChainID > 0 && networkChainID.Int64() != networkConfig.ChainID {
		client.Close()
		return nil, fmt.Errorf("chain id mismatch: configured=%d rpc=%d", networkConfig.ChainID, networkChainID.Int64())
	}

	return &RPC{url: rpcURL, client: client}, nil
}

func safeRPCURLLabel(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Host == "" {
		return "configured endpoint"
	}

	if parsedURL.Scheme == "" {
		return parsedURL.Host
	}
	return parsedURL.Scheme + "://" + parsedURL.Host
}

func (r *RPC) URL() string {
	if r == nil {
		return ""
	}
	return r.url
}

func (r *RPC) Client() *ethclient.Client {
	if r == nil {
		return nil
	}
	return r.client
}

func (r *RPC) Close() {
	if r == nil || r.client == nil {
		return
	}
	r.client.Close()
}
