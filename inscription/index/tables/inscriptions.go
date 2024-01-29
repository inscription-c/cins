package tables

import (
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
	"time"
)

type Inscriptions struct {
	Id              uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"` // this is sequence_num
	InscriptionId   `gorm:"embedded"`
	Index           uint32    `gorm:"column:index;type:int unsigned;default:0;NOT NULL"` // outpoint index of tx
	SequenceNum     uint64    `gorm:"column:sequence_num;type:bigint unsigned;index:idx_sequence_num;default:0;NOT NULL"`
	InscriptionNum  int64     `gorm:"column:inscription_num;type:bigint;index:idx_inscription_num;default:0;NOT NULL"`
	Owner           string    `gorm:"column:owner;type:varchar(255);index:idx_owner;default:;NOT NULL"`
	Charms          uint16    `gorm:"column:charms;type:tinyint unsigned;default:0;NOT NULL"`
	Fee             uint64    `gorm:"column:fee;type:bigint unsigned;default:0;NOT NULL"`
	Height          uint32    `gorm:"column:height;type:int unsigned;default:0;NOT NULL"`
	Sat             uint64    `gorm:"column:sat;type:bigint unsigned;index:idx_sat;default:0;NOT NULL"`
	Timestamp       int64     `gorm:"column:timestamp;type:bigint unsigned;default:0;NOT NULL"`
	Body            []byte    `gorm:"column:body;type:mediumblob;default:;NOT NULL"`
	ContentEncoding string    `gorm:"column:content_encoding;varchar(255);default:;NOT NULL"`
	ContentType     string    `gorm:"column:content_type;varchar(255);default:;NOT NULL"`
	ContentSize     uint32    `gorm:"column:content_size;type:int unsigned;default:0;NOT NULL"`
	ContractDesc    string    `gorm:"column:contract_desc;varchar(255);default:;NOT NULL"`
	Metadata        []byte    `gorm:"column:metadata;type:mediumblob;default:;NOT NULL"`
	Pointer         int32     `gorm:"column:pointer;type:int;default:0;NOT NULL"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (i *Inscriptions) TableName() string {
	return "inscriptions"
}

type InscriptionId struct {
	TxId   string `gorm:"column:tx_id;type:varchar(255);index:idx_tx_id;default:;NOT NULL" json:"txid"` // tx id
	Offset uint32 `gorm:"column:offset;type:int unsigned;default:0;NOT NULL" json:"offset"`             // inscription offset of tx
}

func (i *InscriptionId) MarshalJSON() ([]byte, error) {
	inscriptionId := NewInscriptionId(i.TxId, i.Offset).String()
	return []byte(fmt.Sprintf("\"%s\"", inscriptionId)), nil
}

func (i *InscriptionId) String() string {
	return fmt.Sprintf("%s%s%d", i.TxId, constants.InscriptionIdDelimiter, i.Offset)
}

func NewInscriptionId(txid string, offset uint32) *InscriptionId {
	return &InscriptionId{
		TxId:   txid,
		Offset: offset,
	}
}

func StringToInscriptionId(s string) *InscriptionId {
	s = strings.ToLower(strings.TrimSpace(s))
	if !constants.InscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.InscriptionIdDelimiter)
	return &InscriptionId{
		TxId:   insId[0],
		Offset: gconv.Uint32(insId[1]),
	}
}
