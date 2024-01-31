package inscription

import (
	"github.com/inscription-c/insc/internal/signal"
	"testing"
)

func TestInscribe(t *testing.T) {
	testnet = true
	postage = 1
	inscriptionsFilePath = "./test/cbrc20.json"
	unlockConditionFile = "./test/unlock_condition.json"
	destination = "tb1qq2lsrdnylv0qu7eezsruhv29jxrujm3fpzfpkf"
	if err := inscribe(); err != nil {
		t.Fatal(err)
	}
	<-signal.InterruptHandlersDone
}
