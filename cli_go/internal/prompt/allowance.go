package prompt

import (
	"bufio"
	"dex/internal/amount"
	"fmt"
	"math/big"
	"os"
	"strings"
)

// SelectAllowanceAmount asks the user to choose an allowance amount.
// If requiredAmount is provided, option [1] uses that exact value.
func SelectAllowanceAmount(requiredAmount *big.Int, currentBalance *big.Int, decimals uint8) (*big.Int, error) {
	for {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Select allowance amount:")
		if requiredAmount != nil {
			fmt.Fprintf(os.Stderr, "[1] Order amount (%s)\n", requiredAmount.String())
			if currentBalance != nil {
				fmt.Fprintf(os.Stderr, "[2] Current balance (%s)\n", currentBalance.String())
			} else {
				fmt.Fprintln(os.Stderr, "[2] Current balance (unavailable)")
			}
			fmt.Fprintln(os.Stderr, "[3] Max (uint256)")
			fmt.Fprintln(os.Stderr, "[4] Custom amount")
			fmt.Fprint(os.Stderr, "Select [1/2/3/4]: ")
		} else {
			if currentBalance != nil {
				fmt.Fprintf(os.Stderr, "[1] Current balance (%s)\n", currentBalance.String())
			} else {
				fmt.Fprintln(os.Stderr, "[1] Current balance (unavailable)")
			}
			fmt.Fprintln(os.Stderr, "[2] Max (uint256)")
			fmt.Fprintln(os.Stderr, "[3] Custom amount")
			fmt.Fprint(os.Stderr, "Select [1/2/3]: ")
		}

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("aborted")
		}

		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			if requiredAmount != nil {
				return cloneBigInt(requiredAmount), nil
			}
			if currentBalance == nil {
				fmt.Fprintln(os.Stderr, "Current balance is unavailable.")
				continue
			}
			if currentBalance.Sign() <= 0 {
				fmt.Fprintln(os.Stderr, "Current balance is zero.")
				continue
			}
			return cloneBigInt(currentBalance), nil
		case "2":
			if requiredAmount != nil {
				if currentBalance == nil {
					fmt.Fprintln(os.Stderr, "Current balance is unavailable.")
					continue
				}
				if currentBalance.Sign() <= 0 {
					fmt.Fprintln(os.Stderr, "Current balance is zero.")
					continue
				}
				return cloneBigInt(currentBalance), nil
			}
			return maxUint256(), nil
		case "3":
			if requiredAmount != nil {
				return maxUint256(), nil
			}
			return promptPositiveDecimalAmount("Custom allowance amount", decimals)
		case "4":
			if requiredAmount == nil {
				fmt.Fprintln(os.Stderr, "Please enter 1, 2, or 3.")
				continue
			}
			return promptPositiveDecimalAmount("Custom allowance amount", decimals)
		default:
			if requiredAmount != nil {
				fmt.Fprintln(os.Stderr, "Please enter 1, 2, 3, or 4.")
			} else {
				fmt.Fprintln(os.Stderr, "Please enter 1, 2, or 3.")
			}
		}
	}
}

func promptPositiveDecimalAmount(label string, decimals uint8) (*big.Int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s (decimal token units): ", label)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("aborted")
		}

		value := strings.TrimSpace(scanner.Text())
		n, err := amount.ParseUnits(value, decimals)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid amount %q: %v\n", value, err)
			continue
		}
		if n.Sign() <= 0 {
			fmt.Fprintln(os.Stderr, "Amount must be greater than zero.")
			continue
		}
		return n, nil
	}
}

func maxUint256() *big.Int {
	max := new(big.Int).Lsh(big.NewInt(1), 256)
	return max.Sub(max, big.NewInt(1))
}

func cloneBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
