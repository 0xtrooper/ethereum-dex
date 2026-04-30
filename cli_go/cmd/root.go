package cmd

import (
	"fmt"
	"os"

	configcmd "dex/cmd/config"
	walletcmd "dex/cmd/wallet"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(svc *service.Service, ks *service.Keystore) *cobra.Command {
	root := &cobra.Command{
		Use:   "dex",
		Short: "Decentralized Exchange CLI",
		Long:  `Decentralized Exchange Command Line Interface.`,
	}

	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(configcmd.NewCommand(svc))
	root.AddCommand(walletcmd.NewCommand(ks))

	return root
}

func Execute() {
	svc, err := service.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}

	ks, err := service.NewKeystore()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}

	if err := NewCommand(svc, ks).Execute(); err != nil {
		os.Exit(exitFailure)
	}
}
