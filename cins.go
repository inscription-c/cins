package main

import (
	"github.com/inscription-c/cins/btcd"
	"github.com/inscription-c/cins/inscription"
	"github.com/inscription-c/cins/inscription/server"
	"github.com/inscription-c/cins/wallet"
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
	rootCmd.AddCommand(server.Cmd)
	rootCmd.AddCommand(btcd.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
