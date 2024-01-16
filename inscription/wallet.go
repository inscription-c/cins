package inscription

import "github.com/dotbitHQ/insc/inscription/log"

func OnClientConnected() {
	log.Log.Info("wallet client connected")
}
