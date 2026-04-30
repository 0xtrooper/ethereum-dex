package wallet

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/internal/prompt"
	"dex/service"

	"github.com/spf13/cobra"
	bip39 "github.com/tyler-smith/go-bip39"
)

const walletImportDisclosure = `
WALLET IMPORT - PLEASE READ CAREFULLY BEFORE PROCEEDING

1. The key you provide will be stored locally in an encrypted keystore file on this device.
2. The keystore is protected by the password you set below.
3. Your private key grants unconditional control over any assets held at the associated address.
4. You are solely responsible for key/password security.
5. This software is provided as-is, without warranty of any kind.
`

const walletImportTypeMenu = `
What would you like to import?

[1] Mnemonic phrase (12 words)
[2] Private key (hex)
`

func newWalletImportCommand(ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a wallet into the keystore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKeyFlag, _ := cmd.Flags().GetString("private-key")
			if strings.TrimSpace(privateKeyFlag) != "" {
				password, err := walletImportPassword(cmd)
				if err != nil {
					return err
				}
				address, err := ks.Import(strings.TrimSpace(privateKeyFlag), password)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Imported wallet: %s\n", address)
				return nil
			}

			fmt.Fprint(os.Stderr, walletImportDisclosure)
			agreed, err := prompt.Confirm("I have read and understood the above")
			if err != nil {
				return err
			}
			if !agreed {
				return fmt.Errorf("aborted")
			}

			fmt.Fprint(os.Stderr, walletImportTypeMenu)
			importType, err := walletPromptImportType()
			if err != nil {
				return err
			}

			password, err := walletImportPassword(cmd)
			if err != nil {
				return err
			}

			var address string
			if importType == "mnemonic" {
				mnemonic, err := walletPromptMnemonic()
				if err != nil {
					return err
				}
				address, err = ks.CreateFromMnemonic(mnemonic, password)
				if err != nil {
					return err
				}
			} else {
				privateKey, err := walletPromptPrivateKey()
				if err != nil {
					return err
				}
				address, err = ks.Import(privateKey, password)
				if err != nil {
					return err
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported wallet: %s\n", address)
			return nil
		},
	}

	cmd.Flags().String("password", "", "Keystore password")
	cmd.Flags().String("private-key", "", "Private key hex (skips interactive prompt)")
	return cmd
}

func walletImportPassword(cmd *cobra.Command) (string, error) {
	password, _ := cmd.Flags().GetString("password")
	if password != "" {
		return password, nil
	}
	return prompt.PasswordConfirm("Password to encrypt the imported key")
}

func walletPromptImportType() (string, error) {
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

func walletPromptMnemonic() (string, error) {
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

func walletPromptPrivateKey() (string, error) {
	fmt.Fprintln(os.Stderr, "\nEnter your private key (hex, with or without 0x prefix):")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprint(os.Stderr, "> ")
	scanner.Scan()
	return strings.TrimSpace(scanner.Text()), nil
}
