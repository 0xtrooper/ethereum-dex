// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package amount

import (
	"fmt"
	"math/big"
	"strings"
)

// ParseUnits converts a decimal amount string (human units) into base units.
// Example: ParseUnits("1.5", 18) => 1500000000000000000.
func ParseUnits(value string, decimals uint8) (*big.Int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("amount is required")
	}

	if strings.HasPrefix(value, "+") {
		value = strings.TrimPrefix(value, "+")
	}
	if strings.HasPrefix(value, "-") {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid decimal amount %q", value)
	}

	wholePart := parts[0]
	fractionPart := ""
	if len(parts) == 2 {
		fractionPart = parts[1]
	}
	if wholePart == "" && fractionPart == "" {
		return nil, fmt.Errorf("invalid decimal amount %q", value)
	}
	if wholePart == "" {
		wholePart = "0"
	}

	if !allDigits(wholePart) || (fractionPart != "" && !allDigits(fractionPart)) {
		return nil, fmt.Errorf("invalid decimal amount %q", value)
	}
	if len(fractionPart) > int(decimals) {
		return nil, fmt.Errorf("too many decimal places in %q (max %d)", value, decimals)
	}

	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	whole := new(big.Int)
	if _, ok := whole.SetString(wholePart, 10); !ok {
		return nil, fmt.Errorf("invalid whole amount %q", wholePart)
	}
	whole.Mul(whole, base)

	if fractionPart == "" {
		return whole, nil
	}

	frac := new(big.Int)
	if _, ok := frac.SetString(fractionPart, 10); !ok {
		return nil, fmt.Errorf("invalid fractional amount %q", fractionPart)
	}
	scale := int(decimals) - len(fractionPart)
	if scale > 0 {
		frac.Mul(frac, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	}

	return whole.Add(whole, frac), nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FormatUnits converts a base-unit integer into a decimal string with at least
// one fractional digit when decimals > 0 (e.g. 1e18 => "1.0").
func FormatUnits(raw *big.Int, decimals uint8) string {
	if raw == nil {
		if decimals > 0 {
			return "0.0"
		}
		return "0"
	}
	if decimals == 0 {
		return raw.String()
	}

	sign := ""
	value := new(big.Int).Set(raw)
	if value.Sign() < 0 {
		sign = "-"
		value.Abs(value)
	}

	denom := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Div(value, denom)
	frac := new(big.Int).Mod(value, denom)
	if frac.Sign() == 0 {
		return sign + whole.String() + ".0"
	}

	fracStr := frac.String()
	if len(fracStr) < int(decimals) {
		fracStr = strings.Repeat("0", int(decimals)-len(fracStr)) + fracStr
	}
	fracStr = strings.TrimRight(fracStr, "0")
	if fracStr == "" {
		fracStr = "0"
	}
	return sign + whole.String() + "." + fracStr
}
