package service

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

type Keystore struct {
	ks *keystore.KeyStore
}

func NewKeystore() (*Keystore, error) {
	dir, err := keystoreDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	return &Keystore{ks: ks}, nil
}

func keystoreDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dex", "keystore"), nil
}

func (k *Keystore) Import(privateKeyHex string, password string) (string, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	account, err := k.ks.ImportECDSA(privateKey, password)
	if err != nil {
		return "", err
	}
	return account.Address.Hex(), nil
}

func (k *Keystore) Create(password string) (string, error) {
	account, err := k.ks.NewAccount(password)
	if err != nil {
		return "", err
	}
	return account.Address.Hex(), nil
}

func (k *Keystore) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128) // 12 words
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func (k *Keystore) CreateFromMnemonic(mnemonic string, password string) (string, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic phrase")
	}
	privateKey, err := deriveEthKey(mnemonic)
	if err != nil {
		return "", err
	}
	account, err := k.ks.ImportECDSA(privateKey, password)
	if err != nil {
		return "", err
	}
	return account.Address.Hex(), nil
}

// deriveEthKey derives the Ethereum private key at m/44'/60'/0'/0/0.
func deriveEthKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	path := []uint32{
		bip32.FirstHardenedChild + 44,
		bip32.FirstHardenedChild + 60,
		bip32.FirstHardenedChild + 0,
		0,
		0,
	}
	key := master
	for _, idx := range path {
		key, err = key.NewChildKey(idx)
		if err != nil {
			return nil, err
		}
	}
	return crypto.ToECDSA(key.Key)
}

func (k *Keystore) Export(address string, password string) (string, error) {
	addr := common.HexToAddress(address)
	account, err := k.findAccount(addr)
	if err != nil {
		return "", err
	}
	keyjson, err := k.ks.Export(account, password, password)
	if err != nil {
		return "", err
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(crypto.FromECDSA(key.PrivateKey)), nil
}

func (k *Keystore) List() []string {
	accs := k.ks.Accounts()
	addresses := make([]string, len(accs))
	for i, acc := range accs {
		addresses[i] = acc.Address.Hex()
	}
	return addresses
}

func (k *Keystore) Delete(address string) error {
	addr := common.HexToAddress(address)
	account, err := k.findAccount(addr)
	if err != nil {
		return err
	}
	return os.Remove(account.URL.Path)
}

func (k *Keystore) DeleteAll() error {
	dir, err := keystoreDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (k *Keystore) findAccount(address common.Address) (accounts.Account, error) {
	for _, acc := range k.ks.Accounts() {
		if acc.Address == address {
			return acc, nil
		}
	}
	return accounts.Account{}, fmt.Errorf("account %s not found in keystore", address.Hex())
}
