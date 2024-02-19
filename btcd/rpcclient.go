package btcd

import "github.com/inscription-c/insc/btcd/rpcclient"

func NewClient(host, user, password string) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         password,
		HTTPPostMode: true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewBatchClient(host, user, password string) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         password,
		HTTPPostMode: true,
	}
	return rpcclient.NewBatch(connCfg)
}
