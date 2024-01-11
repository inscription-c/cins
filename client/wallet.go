package client

import (
	"encoding/hex"
	"errors"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	DefaultWalletAccountName = "default"
)

type ListUnspentReq struct {
	Min       uint64
	Max       uint64 `validate:"gtefield=Min"`
	Addresses []string
}

type ListUnspentResp struct {
	OutPoint
	Address       string `json:"address"`
	Account       string `json:"account"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        Amount `json:"amount"`
	Confirmations int    `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
}

func (c *Client) ListUnspent(req *ListUnspentReq) ([]ListUnspentResp, error) {
	if err := validate.Struct(req); err != nil {
		return nil, err
	}
	if req.Max == 0 {
		req.Max = 9999999
	}
	resp := make([]ListUnspentResp, 0)
	if err := c.SendRequest("listunspent", &resp, req.Min, req.Max, req.Addresses); err != nil {
		return nil, err
	}
	return resp, nil
}

type listLockUnspentResp struct {
	OutPoint
}

func (c *Client) ListLockUnspent() ([]listLockUnspentResp, error) {
	resp := make([]listLockUnspentResp, 0)
	if err := c.SendRequest("listlockunspent", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type GetRawTransactionResp struct {
	Hex      string `json:"hex"`
	Txid     string `json:"txid"`
	Hash     string `json:"hash"`
	Size     int    `json:"size"`
	Vsize    int    `json:"vsize"`
	Weight   int    `json:"weight"`
	Version  int    `json:"version"`
	Locktime int    `json:"locktime"`
	Vin      []struct {
		Txid      string `json:"txid"`
		Vout      int    `json:"vout"`
		ScriptSig struct {
			Asm string `json:"asm"`
			Hex string `json:"hex"`
		} `json:"scriptSig"`
		Txinwitness []string `json:"txinwitness"`
		Sequence    int64    `json:"sequence"`
	} `json:"vin"`
	Vout []struct {
		Value        Amount `json:"value"`
		N            int    `json:"n"`
		ScriptPubKey struct {
			Asm       string   `json:"asm"`
			Hex       string   `json:"hex"`
			ReqSigs   int      `json:"reqSigs"`
			Type      string   `json:"type"`
			Address   string   `json:"address"`
			Addresses []string `json:"addresses"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
	Blockhash     string `json:"blockhash"`
	Confirmations int    `json:"confirmations"`
	Time          int    `json:"time"`
	Blocktime     int    `json:"blocktime"`
}

func (c *Client) GetRawTransaction(txid string) (*GetRawTransactionResp, error) {
	resp := &GetRawTransactionResp{}
	if err := c.SendRequest("getrawtransaction", resp, txid, 1); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetChangeAddress(addressType ...string) (string, error) {
	addrTy := ""
	if len(addressType) > 0 && addressType[0] != "" {
		addrTy = addressType[0]
	}
	resp := new(string)
	if err := c.SendRequest("getrawchangeaddress", resp, DefaultWalletAccountName, addrTy); err != nil {
		return "", err
	}
	return *resp, nil
}

type SignRawTransactionWithWalletResp struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
	Errors   []struct {
		Txid      string `json:"txid"`
		Vout      int    `json:"vout"`
		ScriptSig string `json:"scriptSig"`
		Sequence  int    `json:"sequence"`
		Error     string `json:"error"`
	} `json:"errors"`
}

func (c *Client) SignRawTransaction(txRaw []byte) (*SignRawTransactionWithWalletResp, error) {
	resp := new(SignRawTransactionWithWalletResp)
	if err := c.SendRequest("signrawtransaction", resp, hex.EncodeToString(txRaw)); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, errors.New(gconv.String(resp.Errors))
	}
	return resp, nil
}

func (c *Client) WalletPassphrase(password string, unlockSecond int) error {
	if err := c.SendRequest("walletpassphrase", nil, password, unlockSecond); err != nil {
		return err
	}
	return nil
}

func (c *Client) DumpPriKey(address string) (string, error) {
	priKey := new(string)
	if err := c.SendRequest("dumpprivkey", priKey, address); err != nil {
		return "", err
	}
	return *priKey, nil
}
