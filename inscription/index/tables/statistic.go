package tables

import (
	"time"
)

type StatisticType string

const (
	StatisticBlessedInscriptions StatisticType = "BlessedInscriptions"
	StatisticCommits             StatisticType = "Commits"
	StatisticCursedInscriptions  StatisticType = "CursedInscriptions"
	StatisticIndexRunes          StatisticType = "IndexRunes"
	StatisticIndexSats           StatisticType = "IndexSats"
	StatisticLostSats            StatisticType = "LostSats"
	StatisticOutputsTraversed    StatisticType = "OutputsTraversed"
	StatisticReservedRunes       StatisticType = "ReservedRunes"
	StatisticRunes               StatisticType = "Runes"
	StatisticSatRanges           StatisticType = "SatRanges"
	StatisticUnboundInscriptions StatisticType = "UnboundInscriptions"
	StatisticIndexTransactions   StatisticType = "IndexTransactions"
	StatisticIndexSpentSats      StatisticType = "IndexSpentSats"
)

type Statistic struct {
	Id        uint64        `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	Name      StatisticType `gorm:"column:name;type:varchar(255);default:;NOT NULL;comment:statistic name"`
	Count     uint64        `gorm:"column:count;type:bigint unsigned;default:0;NOT NULL;comment:count"`
	CreatedAt time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (s *Statistic) TableName() string {
	return "statistic"
}
