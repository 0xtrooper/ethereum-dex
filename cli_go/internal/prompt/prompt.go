// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
