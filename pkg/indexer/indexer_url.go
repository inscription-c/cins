package indexer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Indexer struct {
	indexerUrl string
}

func NewIndexer(indexerUrl string) *Indexer {
	return &Indexer{indexerUrl: indexerUrl}
}

func (w *Indexer) Outpoint(outpoint string) (*OutpointResp, error) {
	url := fmt.Sprintf("%s/output/%s", w.indexerUrl, outpoint)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	outputResp := &OutpointResp{}

	if err := w.doRetry(request, outputResp); err != nil {
		return nil, err
	}
	return outputResp, nil
}

func (w *Indexer) doRetry(request *http.Request, result interface{}) error {
	idx := 0
	for {
		idx++
		if idx > 1 {
			if idx > 3 {
				return fmt.Errorf("retry 3 times")
			}
			time.Sleep(time.Second)
		}
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			continue
		}
		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()
			return nil
		} else {
			_ = resp.Body.Close()
		}
	}
}
