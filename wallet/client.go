package wallet

import (
	"github.com/btcsuite/btcd/rpcclient"
)

func NewWalletClient(host, user, password string, handlers ...*rpcclient.NotificationHandlers) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:       host,
		Endpoint:   "ws",
		User:       user,
		Pass:       password,
		DisableTLS: true,
	}
	var handler *rpcclient.NotificationHandlers
	if len(handlers) > 0 && handlers[0] != nil {
		handler = handlers[0]
	}
	client, err := rpcclient.New(connCfg, handler)
	if err != nil {
		return nil, err
	}
	return client, nil
}
