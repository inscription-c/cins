package wallet

import (
	"github.com/inscription-c/cins/internal/signal"
	"testing"
)

func TestWallet(t *testing.T) {
	Options.Testnet = true
	Options.IndexSats = "true"
	//indexNoSyncBlock = true
	if err := Main(); err != nil {
		t.Fatal(err)
	}
	<-signal.InterruptHandlersDone
}
