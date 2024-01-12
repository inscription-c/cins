package inscription

import (
	"bytes"
	"io"
)

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
