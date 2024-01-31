package wallet

import (
	"github.com/inscription-c/insc/internal/signal"
	"testing"
)

func TestWallet(t *testing.T) {
	testnet = true
	//indexNoSyncBlock = true
	if err := Main(); err != nil {
		t.Fatal(err)
	}
	<-signal.InterruptHandlersDone
}
