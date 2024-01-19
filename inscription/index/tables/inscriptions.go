package tables

import (
	"time"
)

type Inscriptions struct {
	Id               uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	InscriptionId    string    `gorm:"column:inscription_id;varchar(255);index:idx_inscription_id;default:;NOT NULL"`
	InscriptionNum   uint64    `gorm:"column:inscription_num;type:bigint unsigned;index:idx_inscription_num;default:0;NOT NULL"`
	InscriptionEntry []byte    `gorm:"column:inscription_entry;type:mediumblob;NOT NULL;comment:'inscription entry data'"`
	CreatedAt        time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (i *Inscriptions) TableName() string {
	return "inscriptions"
}
