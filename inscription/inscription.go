package inscription

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/andybalholm/brotli"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/constants"
	"github.com/ugorji/go/codec"
	"io"
	"os"
)

type Inscription struct {
	body []byte

	Header Header
	Body   Protocol
}

type Header struct {
	ChainType       string                `json:"chain_type"`
	ChainId         string                `json:"chain_id"`
	ContentType     constants.ContentType `json:"content_type"`
	ContentEncoding string                `json:"content_encoding"`
	Pointer         string                `json:"pointer"` // TODO pointer
	Metadata        Reader                `json:"metadata"`
}

type Protocol interface {
	Name() string
	Check() error
	Len() int
	Chunks(size uint64) ([]byte, error)
}

type Reader struct {
	data   []byte
	reader *bytes.Reader
}

func (r *Reader) Reset(d []byte) {
	r.data = d
	r.reader = nil
}

func (r *Reader) Data() []byte {
	return r.data
}

func (r *Reader) Len() int {
	if r.data == nil {
		return 0
	}
	return len(r.data)
}

func (r *Reader) Chunks(size uint64) ([]byte, error) {
	if len(r.Data()) == 0 {
		return nil, io.EOF
	}
	if r.reader == nil {
		r.reader = bytes.NewReader(r.data)
	}
	res := make([]byte, size)
	n, err := r.reader.Read(res)
	if err != nil {
		if err == io.EOF {
			if _, err := r.reader.Seek(0, 0); err != nil {
				return nil, err
			}
			return nil, err
		}
		return nil, err
	}
	return res[:n], nil
}

type DefaultProtocol struct {
	Reader
}

func (d *DefaultProtocol) Name() string {
	return "default"
}

func (d *DefaultProtocol) Check() error {
	return nil
}

func FromPath(path string) (*Inscription, error) {
	metadata, err := ParseMetadata(config.CborMetadata, config.JsonMetadata)
	if err != nil {
		return nil, err
	}

	media, err := ContentTypeForPath(path)
	if err != nil {
		return nil, err
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	contentEncoding := ""

	if config.Compress {
		buf := bytes.NewBufferString("")
		bw := brotli.NewWriterOptions(buf, brotli.WriterOptions{Quality: 11, LGWin: 24})
		if _, err := bw.Write(body); err != nil {
			return nil, err
		}
		if err := bw.Close(); err != nil {
			return nil, err
		}
		decompressed := make([]byte, 0)
		if _, err := brotli.NewReader(buf).Read(decompressed); err != nil {
			return nil, err
		}
		if bytes.Compare(body, decompressed) != 0 {
			return nil, errors.New("decompression round trip failed")
		}

		if len(buf.Bytes()) < len(body) {
			body = buf.Bytes()
			contentEncoding = "br"
		}
	}

	var incBody Protocol
	if config.IsBrc20C {
		incBody = &BRC20C{
			DefaultProtocol: DefaultProtocol{
				Reader{
					data: body,
				},
			},
		}
	}
	if incBody == nil {
		incBody = &DefaultProtocol{
			Reader{
				data: body,
			},
		}
	}

	return &Inscription{
		body: body,
		Header: Header{
			ChainType:       config.ChainType,
			ChainId:         config.ChainId,
			ContentType:     media.ContentType,
			ContentEncoding: contentEncoding,
			Pointer:         "",
			Metadata:        Reader{data: metadata},
		},
		Body: incBody,
	}, nil
}

func ParseMetadata(cborMetadata, jsonMetadata string) ([]byte, error) {
	handle := &codec.CborHandle{}
	if cborMetadata != "" {
		data, err := os.ReadFile(cborMetadata)
		if err != nil {
			return nil, err
		}
		var v interface{}
		dec := codec.NewDecoderBytes(data, handle)
		if err := dec.Decode(&v); err != nil {
			return nil, err
		}
		return data, nil
	} else if jsonMetadata != "" {
		data, err := os.ReadFile(jsonMetadata)
		if err != nil {
			return nil, err
		}
		var jsonObj interface{}
		if err := json.Unmarshal(data, &jsonObj); err != nil {
			return nil, err
		}
		cborData := bytes.NewBufferString("")
		codec.NewEncoder(cborData, handle).WriteStr(string(data))
		return cborData.Bytes(), nil
	}
	return nil, nil
}

func (i *Inscription) Data() []byte {
	return i.body
}
