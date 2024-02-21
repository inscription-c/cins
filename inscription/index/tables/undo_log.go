package tables

import "time"

type UndoLog struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Sql       string    `gorm:"column:sql;type:longtext;default:;NOT NULL;comment:sql statement"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (b *UndoLog) TableName() string {
	return "undo_log"
}
