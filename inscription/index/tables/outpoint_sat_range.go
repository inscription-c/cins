package tables

import (
	"time"
)

type OutpointSatRange struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint  string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Start     uint64    `gorm:"column:start;type:bigint unsigned;default:0;NOT NULL;comment:'sat num start'"`
	End       uint64    `gorm:"column:end;type:bigint unsigned;default:0;NOT NULL;comment:'sat num end'"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (o *OutpointSatRange) TableName() string {
	return "outpoint_sat_range"
}
