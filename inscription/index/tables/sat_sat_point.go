package tables

import (
	"time"
)

type SatSatPoint struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Sat       uint64    `gorm:"column:sat;type:bigint unsigned;index:idx_sat;default:0;NOT NULL;comment:sat num"`
	Outpoint  string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset    uint64    `gorm:"column:offset;type:bigint unsigned;default:0;NOT NULL;comment:'offset in sats'"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *SatSatPoint) TableName() string {
	return "sat_sat_point"
}
