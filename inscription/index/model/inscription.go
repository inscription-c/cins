package model

import (
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index/tables"
)

type Inscription struct {
	Body            []byte
	ContentEncoding []byte
	ContentType     constants.ContentType
	CInsDescription tables.CInsDescription
	Metadata        []byte
	Pointer         []byte

	UnRecognizedEvenField bool
	DuplicateField        bool
	IncompleteField       bool
}
