// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package service

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dex/internal/prompt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
}

func NewWallet(ks *Keystore, address string) (*Wallet, error) {
	if ks == nil || ks.ks == nil {
		return nil, fmt.Errorf("keystore service is not initialized")
	}

	account, err := chooseWalletAccount(ks.ks.Accounts(), address)
	if err != nil {
		return nil, err
	}

	password, err := resolveWalletPassword()
	if err != nil {
		return nil, err
	}

	keyJSON, err := os.ReadFile(account.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file for %s: %w", account.Address.Hex(), err)
	}

	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore for %s: %w", account.Address.Hex(), err)
	}

	return &Wallet{
		address:    account.Address,
		privateKey: key.PrivateKey,
	}, nil
}

func (w *Wallet) Address() string {
	if w == nil {
		return ""
	}
	return w.address.Hex()
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	if w == nil {
		return nil
	}
	return w.privateKey
}

func (w *Wallet) PrivateKeyHex() string {
	if w == nil || w.privateKey == nil {
		return ""
	}
	return "0x" + hexEncodePrivateKey(w.privateKey)
}

func chooseWalletAccount(walletAccounts []accounts.Account, requestedAddress string) (accounts.Account, error) {
	if len(walletAccounts) == 0 {
		return accounts.Account{}, fmt.Errorf("no wallets in keystore")
	}

	if requestedAddress != "" {
		trimmedAddress := strings.TrimSpace(requestedAddress)
		if !common.IsHexAddress(trimmedAddress) {
			return accounts.Account{}, fmt.Errorf("invalid wallet address %q", requestedAddress)
		}

		address := common.HexToAddress(trimmedAddress)
		for _, account := range walletAccounts {
			if account.Address == address {
				return account, nil
			}
		}
		return accounts.Account{}, fmt.Errorf("wallet %s not found in keystore", address.Hex())
	}

	if len(walletAccounts) == 1 {
		return walletAccounts[0], nil
	}

	for i, account := range walletAccounts {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, account.Address.Hex())
	}
	fmt.Fprintln(os.Stderr)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "Select wallet [1-%d]: ", len(walletAccounts))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return accounts.Account{}, err
			}
			return accounts.Account{}, fmt.Errorf("aborted")
		}

		selection, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || selection < 1 || selection > len(walletAccounts) {
			fmt.Fprintf(os.Stderr, "Please enter a number between 1 and %d.\n", len(walletAccounts))
			continue
		}
		return walletAccounts[selection-1], nil
	}
}

func resolveWalletPassword() (string, error) {
	password := strings.TrimSpace(os.Getenv(WalletPasswordEnvVar))
	if password != "" {
		return password, nil
	}
	return prompt.Password("Password to decrypt your keystore")
}

func hexEncodePrivateKey(privateKey *ecdsa.PrivateKey) string {
	return fmt.Sprintf("%x", crypto.FromECDSA(privateKey))
}
