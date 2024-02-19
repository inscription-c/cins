package main

import (
	"github.com/inscription-c/insc/btcd"
	"github.com/inscription-c/insc/inscription"
	"github.com/inscription-c/insc/wallet"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cins",
	Short: "cins inscription protocol, include wallet, embed btcd and inscribe tools.",
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
