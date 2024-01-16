package handle

import "github.com/gin-gonic/gin"

type Inscription struct {
	InscriptionId   string `json:"inscription_id"`
	InscriptionNum  uint64 `json:"inscription_number"`
	Next            string `json:"next"`
	Previous        string `json:"previous"`
	Address         string `json:"address"`
	ContentLength   uint64 `json:"content_length"`
	ContentType     string `json:"content_type"`
	GenesisFee      uint64 `json:"genesis_fee"`
	GenesisHeight   uint64 `json:"genesis_height"`
	OutputValue     uint64 `json:"output_value"`
	SatPoint        string `json:"satpoint"`
	Timestamp       uint64 `json:"timestamp"`
	DstChain        string `json:"dst_chain"`
	ContentProtocol string `json:"content_protocol"`
}

func (h *Handler) Inscription(ctx *gin.Context) {
	query := ctx.Param("query")
	if query == "" {
		ctx.JSON(400, gin.H{"error": "query is empty"})
		return
	}

}
