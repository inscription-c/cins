package main

import (
	"github.com/dotbitHQ/insc/btcd"
	"github.com/dotbitHQ/insc/inscription"
	"github.com/dotbitHQ/insc/wallet"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "insc",
	Short: "insc inscription protocol, include wallet, embed btcd and inscribe tools.",
}

func init() {
	rootCmd.AddCommand(wallet.Cmd)
	rootCmd.AddCommand(inscription.Cmd)
	rootCmd.AddCommand(btcd.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
