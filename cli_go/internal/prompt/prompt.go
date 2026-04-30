package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func Password(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func PasswordConfirm(label string) (string, error) {
	for {
		password, err := Password(label)
		if err != nil {
			return "", err
		}
		confirm, err := Password("Confirm " + label)
		if err != nil {
			return "", err
		}
		if password == confirm {
			return password, nil
		}
		fmt.Fprintln(os.Stderr, "Passwords do not match, please try again.")
	}
}

// Confirm prints message and asks the user to type y or n.
// Returns false (and no error) when the user declines.
func Confirm(message string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}
