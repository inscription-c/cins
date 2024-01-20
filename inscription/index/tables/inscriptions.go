package tables

import (
	"time"
)

type Inscriptions struct {
	Id             uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"` // this is sequence_num
	Outpoint       string `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset         uint32 `gorm:"column:offset;type:int unsigned;default:0;NOT NULL"`
	InscriptionNum int64  `gorm:"column:inscription_num;type:bigint;index:idx_inscription_num;default:0;NOT NULL"`

	Charms    uint16 `gorm:"column:charms;type:tinyint unsigned;default:0;NOT NULL"`
	Fee       uint64 `gorm:"column:fee;type:bigint unsigned;default:0;NOT NULL"`
	Height    uint32 `gorm:"column:height;type:int unsigned;default:0;NOT NULL"`
	Sat       uint64 `gorm:"column:sat;type:bigint unsigned;index:idx_sat;default:0;NOT NULL"`
	Timestamp int64  `gorm:"column:timestamp;type:bigint unsigned;default:0;NOT NULL"`

	Body            []byte `gorm:"column:body;type:mediumblob;default:;NOT NULL"`
	ContentEncoding string `gorm:"column:content_encoding;varchar(255);default:;NOT NULL"`
	ContentType     string `gorm:"column:content_type;varchar(255);default:;NOT NULL"`
	DstChain        string `gorm:"column:dst_chain;varchar(255);default:;NOT NULL"`
	Metadata        []byte `gorm:"column:metadata;type:mediumblob;default:;NOT NULL"`
	Pointer         int32  `gorm:"column:pointer;type:int;default:0;NOT NULL"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (i *Inscriptions) TableName() string {
	return "inscriptions"
}
