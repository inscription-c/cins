package tables

import (
	"time"
)

type SatPoint struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint    string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Index       uint32    `gorm:"column:index;type:int unsigned;default:0;NOT NULL"`
	SequenceNum uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *SatPoint) TableName() string {
	return "sat_point"
}
