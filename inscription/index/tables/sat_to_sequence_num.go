package tables

import (
	"time"
)

type SatToSequenceNum struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Sat         uint64    `gorm:"column:sat;type:bigint unsigned;index:idx_sat;NOT NULL;comment:sat number"`
	SequenceNum int64     `gorm:"column:sequence_num;type:bigint;index:idx_sequence_num;default:0;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *SatToSequenceNum) TableName() string {
	return "sat_to_sequence_num"
}
