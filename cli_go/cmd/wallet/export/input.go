package walletexport

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dex/internal/prompt"

	"github.com/spf13/cobra"
)

const exportWarning = `
WARNING — SENSITIVE DATA WILL BE DISPLAYED

Your unencrypted private key is about to be printed to your terminal in plain text. Before proceeding, be aware that:

- Anyone who can view your screen, access your terminal history, read your shell logs, or inspect your clipboard may be able to read the key.
- Exposing your private key may result in the immediate and permanent loss of any funds held at the associated address. There is no way to reverse an unauthorized transfer.
- This software accepts no liability for any loss of funds or data resulting from the exposure or misuse of an exported private key.

Ensure your environment is private and secure before proceeding.
`

type dataIn struct {
	Address  string
	Password string
}

func input(cmd *cobra.Command, args []string, wallets []string) (*dataIn, error) {
	fmt.Fprint(os.Stderr, exportWarning)

	agreed, err := prompt.Confirm("I understand the risks and wish to proceed")
	if err != nil {
		return nil, err
	}
	if !agreed {
		return nil, fmt.Errorf("aborted")
	}

	var address string
	if len(args) > 0 {
		address = strings.TrimSpace(args[0])
	} else {
		address, err = promptSelection(cmd, wallets)
		if err != nil {
			return nil, err
		}
	}

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		password, err = prompt.Password("Password to decrypt your keystore")
		if err != nil {
			return nil, err
		}
	}

	return &dataIn{Address: address, Password: password}, nil
}

func promptSelection(cmd *cobra.Command, wallets []string) (string, error) {
	if len(wallets) == 0 {
		return "", fmt.Errorf("no wallets in keystore")
	}

	for i, addr := range wallets {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d] %s\n", i+1, addr)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "Select wallet to export: ")
		scanner.Scan()
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(wallets) {
			fmt.Fprintf(os.Stderr, "Please enter a number between 1 and %d.\n", len(wallets))
			continue
		}
		return wallets[n-1], nil
	}
}
