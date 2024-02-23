package tables

import (
	"encoding/binary"
	"errors"
	"github.com/inscription-c/cins/internal/math/uint128"
	"time"
)

var ErrInvalidSatRange = errors.New("invalid sat range")

type OutpointSatRange struct {
	Id        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Outpoint  string    `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	SatRange  []byte    `gorm:"column:sat_range;type:longblob;default:;NOT NULL"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (o *OutpointSatRange) TableName() string {
	return "outpoint_sat_range"
}

type SatRanges []*SatRange

type SatRange struct {
	Start uint64
	End   uint64
}

func NewSatRanges(data []byte) (SatRanges, error) {
	res := make(SatRanges, 0, len(data)/11)
	if len(data) == 0 {
		return res, nil
	}
	for i := 0; i < len(data); i += 11 {
		satRange, err := NewSatRange(data[i : i+11])
		if err != nil {
			return nil, err
		}
		res = append(res, satRange)
	}
	return res, nil
}

func NewSatRange(v []byte) (*SatRange, error) {
	if len(v) != 11 {
		return nil, ErrInvalidSatRange
	}
	b0, b1, b2, b3, b4, b5, b6, b7, b8, b9, b10 := v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10]
	rawBase := binary.LittleEndian.Uint64([]byte{b0, b1, b2, b3, b4, b5, b6, 0})
	// 51 bit base
	base := rawBase & ((1 << 51) - 1)
	rawDelta := binary.LittleEndian.Uint64([]byte{b6, b7, b8, b9, b10, 0, 0, 0})
	// 33 bit delta
	delta := rawDelta >> 3
	return &SatRange{
		Start: base,
		End:   base + delta,
	}, nil
}

func (s *SatRange) Store() []byte {
	base := s.Start
	delta := s.End - s.Start

	n := uint128.From64(base).Or(uint128.From64(delta).Lsh(51))
	bs := n.Big().Bytes()
	l := len(bs)
	resLen := 11
	res := make([]byte, resLen)
	for i := l; i > 0; i-- {
		res[l-i] = bs[i-1]
		if l-i+1 == resLen {
			break
		}
	}
	return res
}
