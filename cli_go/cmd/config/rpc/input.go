package rpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type dataIn struct {
	RPCURL string
}

func input(_ *cobra.Command, args []string) (*dataIn, error) {
	rpcURL := strings.TrimSpace(args[0])
	if rpcURL == "" {
		return nil, fmt.Errorf("rpc url cannot be empty")
	}
	return &dataIn{RPCURL: rpcURL}, nil
}
