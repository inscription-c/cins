package wallet

import (
	"github.com/inscription-c/cins/pkg/signal"
	"testing"
)

func TestWallet(t *testing.T) {
	Options.Testnet = true
	//indexNoSyncBlock = true
	if err := Main(); err != nil {
		t.Fatal(err)
	}
	<-signal.InterruptHandlersDone
}
