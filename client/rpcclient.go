package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dotbitHQ/insc/config"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/btcsuite/btcd/btcjson"
)

var validate = validator.New()

type Client struct {
	url      string
	user     string
	password string

	tLSSkipVerify bool
	cert          string
}

var clientLock sync.Once
var client *Client

func RPC() *Client {
	clientLock.Do(func() {
		client = &Client{
			url:           config.RpcConnect,
			user:          config.Username,
			password:      config.Password,
			tLSSkipVerify: config.TLSSkipVerify,
			cert:          config.RPCCert,
		}
	})
	return client
}

func (c *Client) SendRequest(method string, result interface{}, params ...interface{}) error {
	cmd, err := btcjson.NewCmd(method, params...)
	if err != nil {
		var jerr btcjson.Error
		if errors.As(err, &jerr) {
			return fmt.Errorf("%s command: %v (code: %s)\n", method, err, jerr.ErrorCode)
		}
		return fmt.Errorf("%s command: %v\n", method, err)
	}

	marshalledJSON, err := btcjson.MarshalCmd(btcjson.RpcVersion2, 1, cmd)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(marshalledJSON)
	httpRequest, err := http.NewRequest("POST", c.url, bodyReader)
	if err != nil {
		return err
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(config.Username, config.Password)
	httpClient, err := c.newHTTPClient()
	if err != nil {
		return err
	}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return err
	}

	respBytes, err := io.ReadAll(httpResponse.Body)
	_ = httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return err
	}

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		if len(respBytes) == 0 {
			return fmt.Errorf("%d %s", httpResponse.StatusCode, http.StatusText(httpResponse.StatusCode))
		}
		return fmt.Errorf("%s", respBytes)
	}
	resp := &Response{
		Result: result,
	}
	if err := json.Unmarshal(respBytes, resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func (c *Client) newHTTPClient() (*http.Client, error) {
	var tlsConfig *tls.Config
	if c.cert != "" {
		pem, err := os.ReadFile(c.cert)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsConfig = &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: c.tLSSkipVerify,
		}
	}

	client := http.Client{}
	if tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}
	return &client, nil
}
