package chainid

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type dataIn struct {
	ChainID int64
}

func input(_ *cobra.Command, args []string) (*dataIn, error) {
	rawChainID := strings.TrimSpace(args[0])
	chainID, err := strconv.ParseInt(rawChainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chain id %q: %w", rawChainID, err)
	}
	if chainID <= 0 {
		return nil, fmt.Errorf("chain id must be greater than zero")
	}
	return &dataIn{ChainID: chainID}, nil
}
