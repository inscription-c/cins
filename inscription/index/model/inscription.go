package model

type Inscription struct {
	Body            []byte
	ContentEncoding []byte
	ContentType     []byte
	DstChain        []byte
	Metadata        []byte
	Pointer         []byte

	UnRecognizedEvenField bool
	DuplicateField        bool
	IncompleteField       bool
}
