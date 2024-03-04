package indexer

type IndexerInterface interface {
	Outpoint(outpoint string) (*OutpointResp, error)
}

type OutpointResp struct {
	Address      string   `json:"address"`
	Inscriptions []string `json:"inscriptions"`
	ScriptPubKey string   `json:"script_pubkey"`
	Transaction  string   `json:"transaction"`
	Value        int64    `json:"value"`
}
