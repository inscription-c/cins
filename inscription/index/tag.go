package index

import (
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
)

type TagType int

const (
	TagPointer TagType = iota
	TagUnbound

	TagContentType
	TagParent
	TagMetadata
	TagMetaprotocol
	TagContentEncoding
	TagDelegate
	TagDstChain

	TagNop
)

func TagFromBytes(bs []byte) TagType {
	switch bs[0] {
	case 2:
		return TagPointer
	case 66:
		return TagUnbound

	case 1:
		return TagContentType
	case 3:
		return TagParent
	case 5:
		return TagMetadata
	case 7:
		return TagMetaprotocol
	case 9:
		return TagContentEncoding
	case 11:
		return TagDelegate
	case 255:
		return TagNop
	default:
		if string(bs) == gconv.String(constants.DstChain) {
			return TagDstChain
		}
		return TagNop
	}
}

func (t TagType) Bytes() []byte {
	switch t {
	case TagPointer:
		return []byte{2}
	case TagUnbound:
		return []byte{66}

	case TagContentType:
		return []byte{1}
	case TagParent:
		return []byte{3}
	case TagMetadata:
		return []byte{5}
	case TagMetaprotocol:
		return []byte{7}
	case TagContentEncoding:
		return []byte{9}
	case TagDelegate:
		return []byte{11}
	case TagNop:
		return []byte{255}
	case TagDstChain:
		return []byte(gconv.String(constants.DstChain))
	default:
		return []byte{}
	}
}

func (t TagType) IsChunked() bool {
	return t == TagMetadata
}

func (t TagType) RemoveField(fields map[TagType][][]byte) []byte {
	if t.IsChunked() {
		value, ok := fields[t]
		if !ok {
			return []byte{}
		} else {
			delete(fields, t)
			var res []byte
			for _, v := range value {
				res = append(res, v...)
			}
			return res
		}
	} else {
		values, ok := fields[t]
		if !ok {
			return []byte{}
		} else {
			var res []byte
			if len(values) > 0 {
				res = values[0]
				values = values[1:]
			}
			if len(values) == 0 {
				delete(fields, t)
			}
			return res
		}
	}
}
