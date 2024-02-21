package tables

import "time"

type SavePoint struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Height    uint32    `gorm:"column:height;type:int unsigned;index:idx_height;default:0;NOT NULL"`
	UndoLogId uint64    `gorm:"column:undo_log_id;type:bigint unsigned;index:idx_undo_log_id;default:0;NOT NULL"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (b *SavePoint) TableName() string {
	return "savepoint"
}
