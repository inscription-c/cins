package indexer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Indexer struct {
	indexerUrl string
}

type OutpointResp struct {
	Address      string   `json:"address"`
	Inscriptions []string `json:"inscriptions"`
	ScriptPubKey string   `json:"script_pubkey"`
	Transaction  string   `json:"transaction"`
	Value        int64    `json:"value"`
}

func NewIndexer(indexerUrl string) *Indexer {
	return &Indexer{indexerUrl: indexerUrl}
}

func (w *Indexer) Outpoint(outpoint string) (*OutpointResp, error) {
	url := fmt.Sprintf("%s/output/%s", w.indexerUrl, outpoint)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	outputResp := &OutpointResp{}
	if err := json.NewDecoder(resp.Body).Decode(outputResp); err != nil {
		return nil, err
	}
	return outputResp, nil
}
