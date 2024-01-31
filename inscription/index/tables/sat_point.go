package tables

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/inscription-c/insc/constants"
	"strconv"
	"strings"
	"time"
)

type SatPoint struct {
	Id          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint    string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset      uint64    `gorm:"column:offset;type:bigint unsigned;default:0;NOT NULL"`
	SequenceNum uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *SatPoint) TableName() string {
	return "sat_point"
}

func (s *SatPoint) String() string {
	return fmt.Sprintf("%s%s%d", s.Outpoint, constants.OutpointDelimiter, s.Offset)
}

func FormatOutpoint(txid string, index uint32) string {
	return fmt.Sprintf("%s%s%d", txid, constants.OutpointDelimiter, index)
}

func FormatSatPoint(outpoint string, sat uint64) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.OutpointDelimiter, sat)
}

func NewSatPointFromString(satpoint string) (*SatPoint, error) {
	parts := strings.Split(satpoint, constants.OutpointDelimiter)
	if len(parts) != 3 {
		return nil, errors.New("satpoint should be of the form txid:index:offset")
	}
	hash, err := chainhash.NewHashFromStr(parts[0])
	if err != nil {
		return nil, err
	}
	outputIndex, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid output index: %v", err)
	}
	offset, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid satpoint offset: %v", err)
	}
	return &SatPoint{
		Outpoint: fmt.Sprintf("%s%s%d", *hash, constants.OutpointDelimiter, outputIndex),
		Offset:   offset,
	}, nil
}
