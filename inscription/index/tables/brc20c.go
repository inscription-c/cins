package tables

import (
	"time"
)

type BRC20C struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"` // this is sequence_num
	Outpoint    string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	SequenceNum uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	Protocol    string    `gorm:"column:protocol;type:varchar(255);index:idx_protocol;default:;NOT NULL"`
	Operator    string    `gorm:"column:operator;type:varchar(255);index:idx_operator;default:;NOT NULL"`
	Ticker      string    `gorm:"column:ticker;type:varchar(255);index:idx_ticker;default:;NOT NULL"`
	Metadata    []byte    `gorm:"column:metadata;type:mediumblob;default:;NOT NULL"`
	Max         uint64    `gorm:"column:max;type:bigint unsigned;default:0;NOT NULL"`
	Limit       uint64    `gorm:"column:limit;type:bigint unsigned;default:0;NOT NULL"`
	Decimal     uint32    `gorm:"column:decimal;type:int unsigned;default:0;NOT NULL"`
	Amount      uint64    `gorm:"column:amount;type:bigint unsigned;default:0;NOT NULL"`
	To          string    `gorm:"column:to;type:varchar(255);index:idx_to;default:;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (t *BRC20C) TableName() string {
	return "brc20c"
}
