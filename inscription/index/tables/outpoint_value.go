package tables

import (
	"time"
)

type OutpointValue struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint  string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Value     int64     `gorm:"column:value;type:bigint unsigned;default:0;NOT NULL;comment:'value on outpoint'"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (o *OutpointValue) TableName() string {
	return "outpoint_value"
}
