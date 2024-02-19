package handle

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SearchType string

const (
	SearchTypeUnknown           SearchType = "unknown"
	SearchTypeEmpty             SearchType = "empty"
	SearchTypeInscriptionId     SearchType = "inscription_id"
	SearchTypeInscriptionNumber SearchType = "inscription_number"
	SearchTypeAddress           SearchType = "address"
	//SearchTypeProtocol          SearchType = "protocol"
)

type ScanInscriptionsReq struct {
	Search          string   `json:"search"`
	Page            int      `json:"page" binding:"omitempty,min=1"`
	Limit           int      `json:"limit" binding:"omitempty,min=1,max=50"`
	Order           string   `json:"order" binding:"omitempty,oneof=newest oldest"`
	Types           []string `json:"types" binding:"omitempty,dive,oneof=image text json html"`
	InscriptionType string   `json:"inscription_type" binding:"omitempty,oneof=c-brc-20"`
}

func (req *ScanInscriptionsReq) Check() error {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Order == "" {
		req.Order = "newest"
	}
	return nil
}

type ScanInscriptionsResp struct {
	SearchType SearchType              `json:"search_type"`
	Page       int                     `json:"page"`
	Total      int                     `json:"total"`
	List       []*ScanInscriptionEntry `json:"list"`
}

type ScanInscriptionEntry struct {
	InscriptionId     string `json:"inscription_id"`
	InscriptionNumber int64  `json:"inscription_number"`
	ContentType       string `json:"content_type"`
	ContentLength     uint32 `json:"content_length"`
	Timestamp         string `json:"timestamp"`
	OwnerOutput       string `json:"owner_output"`
	OwnerAddress      string `json:"owner_address"`
	//Sat               string                 `json:"sat"`
	CInsDescription tables.CInsDescription `json:"c_ins_description"`
	ContentProtocol string                 `json:"content_protocol"`
}

func (h *Handler) ScanInscriptions(ctx *gin.Context) {
	req := &ScanInscriptionsReq{}
	if err := ctx.BindJSON(req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Check(); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.doScanInscriptions(ctx, req); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doScanInscriptions(ctx *gin.Context, req *ScanInscriptionsReq) error {
	resp := &ScanInscriptionsResp{
		SearchType: SearchTypeUnknown,
		Page:       req.Page,
		List:       make([]*ScanInscriptionEntry, 0),
	}

	searParams := &dao.FindProtocolsParams{
		Page:            req.Page,
		Size:            req.Limit,
		Order:           req.Order,
		Types:           req.Types,
		InscriptionType: req.InscriptionType,
	}

	req.Search = strings.TrimSpace(req.Search)
	if req.Search == "" {
		resp.SearchType = SearchTypeEmpty
	}
	if req.Search != "" {
		if constants.InscriptionIdRegexp.MatchString(req.Search) {
			resp.SearchType = SearchTypeInscriptionId
			insId := tables.StringToInscriptionId(req.Search)
			searParams.TxId, searParams.Offset = insId.TxId, insId.Offset
		} else {
			addressAndProtocol := func() {
				if _, err := btcutil.DecodeAddress(req.Search, util.ActiveNet.Params); err == nil {
					resp.SearchType = SearchTypeAddress
					searParams.Owner = req.Search
				}
			}
			inscriptionNumber, err := strconv.ParseInt(req.Search, 10, 64)
			if err == nil {
				inscription, err := h.DB().GetInscriptionByInscriptionNum(inscriptionNumber)
				if err != nil {
					return err
				}
				if inscription.Id > 0 {
					resp.SearchType = SearchTypeInscriptionNumber
					searParams.InscriptionNum = &inscriptionNumber
				} else {
					addressAndProtocol()
				}
			} else {
				addressAndProtocol()
			}
		}
	}

	list, total, err := h.DB().SearchInscriptions(searParams)
	if err != nil {
		return err
	}
	resp.Total = int(total)

	for _, ins := range list {
		resp.List = append(resp.List, &ScanInscriptionEntry{
			InscriptionId:     ins.InscriptionId.String(),
			InscriptionNumber: ins.InscriptionNum,
			ContentType:       ins.MediaType,
			ContentLength:     ins.ContentSize,
			Timestamp:         time.Unix(ins.Timestamp, 0).UTC().Format(time.RFC3339),
			OwnerOutput:       model.NewOutPoint(ins.TxId, ins.Index).String(),
			OwnerAddress:      ins.Owner,
			//Sat:               gconv.String(ins.Sat),
			CInsDescription: ins.CInsDescription,
			ContentProtocol: ins.ContentProtocol,
		})
	}

	ctx.JSON(http.StatusOK, resp)
	return nil
}
