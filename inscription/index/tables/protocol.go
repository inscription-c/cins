package tables

import (
	"time"
)

type Protocol struct {
	Id            uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"` // this is sequence_num
	InscriptionId `gorm:"embedded"`
	Index         uint32    `gorm:"column:index;type:int unsigned;default:0;NOT NULL"` // outpoint index of tx
	SequenceNum   uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	Protocol      string    `gorm:"column:protocol;type:varchar(255);index:idx_protocol;default:;NOT NULL"`
	Ticker        string    `gorm:"column:ticker;type:varchar(255);index:idx_ticker;default:;NOT NULL"`
	Operator      string    `gorm:"column:operator;type:varchar(255);index:idx_operator;default:;NOT NULL"`
	Max           uint64    `gorm:"column:max;type:bigint unsigned;default:0;NOT NULL"`
	Limit         uint64    `gorm:"column:limit;type:bigint unsigned;default:0;NOT NULL"`
	Decimals      uint32    `gorm:"column:decimals;type:int unsigned;default:0;NOT NULL"`
	TkID          string    `gorm:"column:tkid;type:varchar(255);index:idx_tkid;default:;NOT NULL"`
	Amount        uint64    `gorm:"column:amount;type:bigint unsigned;default:0;NOT NULL"`
	To            string    `gorm:"column:to;type:varchar(255);index:idx_to;default:;NOT NULL"`
	Miner         string    `gorm:"column:miner;type:varchar(255);index:idx_miner;default:;NOT NULL"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

// ProtocolAmount is a struct that represents the amount of a protocol.
type ProtocolAmount struct {
	TkID   string `gorm:"column:tkid" json:"ticker_id"`
	Ticker string `gorm:"column:ticker" json:"ticker"`
	Amount uint64 `gorm:"column:amount" json:"amount"`
}

func (t *Protocol) TableName() string {
	return "protocol"
}
