package walletimport

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/internal/prompt"

	"github.com/spf13/cobra"
	bip39 "github.com/tyler-smith/go-bip39"
)

const importDisclosure = `
WALLET IMPORT — PLEASE READ CAREFULLY BEFORE PROCEEDING

1. The key you provide will be stored locally in an encrypted keystore file on this device. It is never transmitted, uploaded, or shared with any third party.

2. The keystore is protected by the password you set below. Without the correct password the key cannot be decrypted or used. There is no password reset or recovery mechanism.

3. Your private key grants unconditional control over any digital assets held at the associated address. Any party who obtains your private key can transfer those assets without restriction and without your consent. Treat it like cash — whoever holds it, owns it.

4. You are solely responsible for the security of the key you are importing and the password you choose to protect it. This software provides no backup, escrow, or recovery service.

5. This software is provided as-is, without warranty of any kind. The authors accept no liability for any loss of funds or data arising from the use, misuse, or compromise of this tool or the keystore it manages.
`

const importTypeMenu = `
What would you like to import?

[1] Mnemonic phrase (12 words)
    Choose this if you have a recovery phrase — a sequence of 12 words you wrote down when
    you first created your wallet. Most wallet apps (MetaMask, Trust Wallet, Ledger, etc.)
    use this format for backup and recovery.

[2] Private key (hex)
    Choose this if you have a raw private key — a 64-character hexadecimal string.
    This is typically exported directly from another tool or wallet.
`

type dataIn struct {
	PrivateKey string // set when importing by hex
	Mnemonic   string // set when importing by mnemonic
	Password   string
}

func input(cmd *cobra.Command, _ []string) (*dataIn, error) {
	// Fast track: --private-key flag bypasses disclosure and choice prompt.
	privateKeyFlag, _ := cmd.Flags().GetString("private-key")
	if privateKeyFlag != "" {
		password, err := getPassword(cmd)
		if err != nil {
			return nil, err
		}
		return &dataIn{PrivateKey: strings.TrimSpace(privateKeyFlag), Password: password}, nil
	}

	fmt.Fprint(os.Stderr, importDisclosure)
	agreed, err := prompt.Confirm("I have read and understood the above")
	if err != nil {
		return nil, err
	}
	if !agreed {
		return nil, fmt.Errorf("aborted")
	}

	fmt.Fprint(os.Stderr, importTypeMenu)
	importType, err := promptImportType()
	if err != nil {
		return nil, err
	}

	password, err := getPassword(cmd)
	if err != nil {
		return nil, err
	}

	switch importType {
	case "mnemonic":
		mnemonic, err := promptMnemonic()
		if err != nil {
			return nil, err
		}
		return &dataIn{Mnemonic: mnemonic, Password: password}, nil
	default:
		privateKey, err := promptPrivateKey()
		if err != nil {
			return nil, err
		}
		return &dataIn{PrivateKey: privateKey, Password: password}, nil
	}
}

func promptImportType() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "Select [1/2]: ")
		scanner.Scan()
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			return "mnemonic", nil
		case "2":
			return "privatekey", nil
		default:
			fmt.Fprintln(os.Stderr, "Please enter 1 or 2.")
		}
	}
}

func promptMnemonic() (string, error) {
	fmt.Fprintln(os.Stderr, "\nEnter your 12-word mnemonic phrase (words separated by spaces):")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "> ")
		scanner.Scan()
		mnemonic := strings.Join(strings.Fields(scanner.Text()), " ")
		if bip39.IsMnemonicValid(mnemonic) {
			return mnemonic, nil
		}
		fmt.Fprintln(os.Stderr, "Invalid mnemonic phrase. Please check your words and try again.")
	}
}

func promptPrivateKey() (string, error) {
	fmt.Fprintln(os.Stderr, "\nEnter your private key (hex, with or without 0x prefix):")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprint(os.Stderr, "> ")
	scanner.Scan()
	return strings.TrimSpace(scanner.Text()), nil
}

func getPassword(cmd *cobra.Command) (string, error) {
	password, _ := cmd.Flags().GetString("password")
	if password != "" {
		return password, nil
	}
	return prompt.PasswordConfirm("Password to encrypt the imported key")
}
