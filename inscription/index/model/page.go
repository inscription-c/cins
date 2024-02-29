package model

type PageResponse struct {
	PageIndex int  `json:"page_index"`
	More      bool `json:"more"`
}
