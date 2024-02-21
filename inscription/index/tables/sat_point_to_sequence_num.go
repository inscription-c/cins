package tables

import (
	"fmt"
	"github.com/inscription-c/insc/constants"
	"time"
)

type SatPointToSequenceNum struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint    string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset      uint64    `gorm:"column:offset;type:bigint unsigned;default:0;NOT NULL"`
	SequenceNum int64     `gorm:"column:sequence_num;type:bigint;index:idx_sequence_num;default:0;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *SatPointToSequenceNum) TableName() string {
	return "sat_point_to_sequence_num"
}

func (s *SatPointToSequenceNum) String() string {
	return fmt.Sprintf("%s%s%d", s.Outpoint, constants.OutpointDelimiter, s.Offset)
}

func FormatOutpoint(txid string, index uint32) string {
	return fmt.Sprintf("%s%s%d", txid, constants.OutpointDelimiter, index)
}

func FormatSatPoint(outpoint string, sat uint64) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.OutpointDelimiter, sat)
}
