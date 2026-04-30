package walletcreate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/internal/prompt"

	"github.com/spf13/cobra"
)

const createDisclosure = `
WALLET CREATION — PLEASE READ CAREFULLY BEFORE PROCEEDING

1. A new private key will be generated and stored locally in an encrypted keystore file on this device. It is never transmitted, uploaded, or shared with any third party.

2. The keystore is protected by the password you set below. Without the correct password the key cannot be decrypted or used. There is no password reset or recovery mechanism.

3. Your private key grants unconditional control over any digital assets held at the associated address. Any party who obtains your private key can transfer those assets without restriction and without your consent.

4. You are solely responsible for keeping your private key and password secure. Loss of your password — or of the keystore file itself — means permanent, irrecoverable loss of access to your funds. This software provides no backup or escrow service.

5. This software is provided as-is, without warranty of any kind. The authors accept no liability for any loss of funds or data arising from the use, misuse, or compromise of this tool or the keystore it manages.
`

type dataIn struct {
	Password string
	Mnemonic string // empty when --no-mnemonic
}

func input(cmd *cobra.Command, ks WalletCreator, _ []string) (*dataIn, error) {
	fmt.Fprint(os.Stderr, createDisclosure)

	agreed, err := prompt.Confirm("I have read and understood the above")
	if err != nil {
		return nil, err
	}
	if !agreed {
		return nil, fmt.Errorf("aborted")
	}

	noMnemonic, _ := cmd.Flags().GetBool("no-mnemonic")

	var mnemonic string
	if !noMnemonic {
		mnemonic, err = mnemonicFlow(ks)
		if err != nil {
			return nil, err
		}
	}

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		password, err = prompt.PasswordConfirm("Password to encrypt your wallet")
		if err != nil {
			return nil, err
		}
	}

	return &dataIn{Password: password, Mnemonic: mnemonic}, nil
}

func mnemonicFlow(ks WalletCreator) (string, error) {
	mnemonic, err := ks.GenerateMnemonic()
	if err != nil {
		return "", err
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Your mnemonic phrase (12 words) — write these down and store them somewhere safe:")
	fmt.Fprintln(os.Stderr)
	words := strings.Fields(mnemonic)
	for i, word := range words {
		fmt.Fprintf(os.Stderr, "  %2d. %s\n", i+1, word)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "This phrase can restore your wallet. Anyone who has it has full access to your funds.")
	fmt.Fprintln(os.Stderr)

	agreed, err := prompt.Confirm("I have written down my mnemonic phrase")
	if err != nil {
		return "", err
	}
	if !agreed {
		return "", fmt.Errorf("aborted")
	}

	// Clear the screen so the mnemonic can no longer be read from scroll-back.
	fmt.Fprint(os.Stderr, "\033[2J\033[H")

	if err := confirmMnemonic(mnemonic); err != nil {
		return "", err
	}

	return mnemonic, nil
}

func confirmMnemonic(expected string) error {
	fmt.Fprintln(os.Stderr, "Please re-enter your mnemonic phrase to confirm you have saved it:")
	fmt.Fprintln(os.Stderr)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "> ")
		scanner.Scan()
		entered := strings.Join(strings.Fields(scanner.Text()), " ")
		if entered == strings.Join(strings.Fields(expected), " ") {
			fmt.Fprintln(os.Stderr)
			return nil
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Mnemonic does not match. Please try again.")
		fmt.Fprintln(os.Stderr)
	}
}
