package btcd

import "github.com/btcsuite/btcd/rpcclient"

func NewClient(host, user, password string, disableTls bool, handlers ...*rpcclient.NotificationHandlers) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:       host,
		Endpoint:   "ws",
		User:       user,
		Pass:       password,
		DisableTLS: disableTls,
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

func NewBatchClient(host, user, password string, disableTls bool) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         password,
		DisableTLS:   disableTls,
		HTTPPostMode: true,
	}
	return rpcclient.NewBatch(connCfg)
}
