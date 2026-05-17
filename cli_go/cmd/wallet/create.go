// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package wallet

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dex/internal/prompt"
	"dex/service"

	"github.com/spf13/cobra"
)

const walletCreateDisclosure = `
WALLET CREATION - PLEASE READ CAREFULLY BEFORE PROCEEDING

1. A new private key will be generated and stored locally in an encrypted keystore file on this device.
2. The keystore is protected by the password you set below.
3. Your private key grants unconditional control over any assets held at the associated address.
4. You are solely responsible for keeping your private key and password secure.
5. This software is provided as-is, without warranty of any kind.
`

func newWalletCreateCommand(ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				fmt.Fprint(os.Stderr, walletCreateDisclosure)
				agreed, err := prompt.Confirm("I have read and understood the above")
				if err != nil {
					return err
				}
				if !agreed {
					return fmt.Errorf("aborted")
				}
			}

			noMnemonic, _ := cmd.Flags().GetBool("no-mnemonic")

			var mnemonic string
			if !noMnemonic {
				mnemonic, err = walletCreateMnemonicFlow(ks, yes)
				if err != nil {
					return err
				}
			}

			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				password, err = prompt.PasswordConfirm("Password to encrypt your wallet")
				if err != nil {
					return err
				}
			}

			var address string
			if mnemonic != "" {
				address, err = ks.CreateFromMnemonic(mnemonic, password)
			} else {
				address, err = ks.Create(password)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created wallet: %s\n", address)
			return nil
		},
	}

	cmd.Flags().String("password", "", "Keystore password")
	cmd.Flags().Bool("no-mnemonic", false, "Skip mnemonic phrase generation and confirmation")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	return cmd
}

func walletCreateMnemonicFlow(ks *service.Keystore, yes bool) (string, error) {
	mnemonic, err := ks.GenerateMnemonic()
	if err != nil {
		return "", err
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Your mnemonic phrase (12 words) - write these down and store them somewhere safe:")
	fmt.Fprintln(os.Stderr)
	words := strings.Fields(mnemonic)
	for i, word := range words {
		fmt.Fprintf(os.Stderr, "  %2d. %s\n", i+1, word)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "This phrase can restore your wallet. Anyone who has it has full access to your funds.")
	fmt.Fprintln(os.Stderr)

	if !yes {
		agreed, err := prompt.Confirm("I have written down my mnemonic phrase")
		if err != nil {
			return "", err
		}
		if !agreed {
			return "", fmt.Errorf("aborted")
		}

		fmt.Fprint(os.Stderr, "\033[2J\033[H")
		if err := walletConfirmMnemonic(mnemonic); err != nil {
			return "", err
		}
	}
	return mnemonic, nil
}

func walletConfirmMnemonic(expected string) error {
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
