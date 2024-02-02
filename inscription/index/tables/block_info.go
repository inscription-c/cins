package tables

import (
	"bytes"
	"github.com/btcsuite/btcd/wire"
	"time"
)

type BlockInfo struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Height      uint32    `gorm:"column:height;type:int unsigned;index:idx_height;default:0;NOT NULL"`
	SequenceNum uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	Header      []byte    `gorm:"column:header;type:blob;NOT NULL;comment:header"`
	Timestamp   int64     `gorm:"column:timestamp;type:bigint;default:0;NOT NULL;comment:timestamp"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (b *BlockInfo) TableName() string {
	return "block_info"
}

func (b *BlockInfo) LoadHeader() (*wire.BlockHeader, error) {
	h := &wire.BlockHeader{}
	buf := bytes.NewBuffer(b.Header)
	if err := h.Deserialize(buf); err != nil {
		return nil, err
	}
	return h, nil
}
