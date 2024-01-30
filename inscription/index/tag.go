package index

import (
	"github.com/inscription-c/insc/constants"
)

// TagType represents the type of tag in the blockchain.
type TagType int

// Constants representing the different types of tags.
const (
	TagPointer TagType = iota
	TagUnbound

	TagContentType
	TagParent
	TagMetadata
	TagMetaprotocol
	TagContentEncoding
	TagDelegate
	TagUnlockCondition

	TagNop
)

// TagFromBytes creates a new TagType from a given byte slice.
// It determines the TagType by comparing the first byte of the slice with the constants.
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
		if string(bs) == constants.UnlockCondition {
			return TagUnlockCondition
		}
		return TagNop
	}
}

// Bytes returns the byte representation of the TagType.
// It determines the byte representation by comparing the TagType with the constants.
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
	case TagUnlockCondition:
		return []byte(constants.UnlockCondition)
	default:
		return []byte{}
	}
}

// IsChunked checks if the TagType is chunked.
// It returns true if the TagType is TagMetadata, false otherwise.
func (t TagType) IsChunked() bool {
	return t == TagMetadata
}

// RemoveField removes a field from a given map of fields.
// If the TagType is chunked, it removes all values associated with the TagType.
// If the TagType is not chunked, it removes the first value associated with the TagType.
// It returns the removed value(s) as a byte slice.
func (t TagType) RemoveField(fields map[TagType][][]byte) []byte {
	var res []byte
	if t.IsChunked() {
		value, ok := fields[t]
		if !ok {
			return res
		}
		delete(fields, t)
		for _, v := range value {
			res = append(res, v...)
		}
		return res
	} else {
		values, ok := fields[t]
		if !ok {
			return res
		}
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
