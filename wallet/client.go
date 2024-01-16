package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

type RpcResp struct {
	JsonRpc string            `json:"jsonrpc"`
	Result  json.RawMessage   `json:"result"`
	Err     *btcjson.RPCError `json:"error"`
	Id      *float64          `json:"id"`
}

// FutureBatchGetRawTransactionResult is a future promise to deliver the result of a
// GetRawTransactionAsync RPC invocation (or an applicable error).
type FutureBatchGetRawTransactionResult chan *rpcclient.Response

// Receive waits for the Response promised by the future and returns a
// transaction given its hash.
func (r FutureBatchGetRawTransactionResult) Receive() ([]*btcutil.Tx, error) {
	res, err := rpcclient.ReceiveFuture(r)
	if err != nil {
		return nil, err
	}
	batchTxResp := make([]*RpcResp, 0)
	err = json.Unmarshal(res, &batchTxResp)
	if err != nil {
		return nil, err
	}
	resp := make([]*btcutil.Tx, 0, len(batchTxResp))
	for _, v := range batchTxResp {
		serializedTx, err := hex.DecodeString(string(v.Result))
		if err != nil {
			return nil, err
		}
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
			return nil, err
		}
		resp = append(resp, btcutil.NewTx(&msgTx))
	}
	return resp, nil
}

func NewWalletClient(host, user, password string, disableTls bool, handlers ...*rpcclient.NotificationHandlers) (*rpcclient.Client, error) {
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
