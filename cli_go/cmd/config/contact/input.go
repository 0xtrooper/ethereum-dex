package contact

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

type dataIn struct {
	Address string
}

func input(_ *cobra.Command, args []string) (*dataIn, error) {
	address := strings.TrimSpace(args[0])
	if !ethAddressPattern.MatchString(address) {
		return nil, fmt.Errorf("invalid Ethereum address %q", address)
	}
	return &dataIn{Address: address}, nil
}
