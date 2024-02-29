package index

import (
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index/dao"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/internal/util"
)

type Protocol struct {
	wtx   *dao.DB
	entry *tables.Inscriptions
}

// NewProtocol is a function that returns a new protocol.
func NewProtocol(wtx *dao.DB, entry *tables.Inscriptions) *Protocol {
	return &Protocol{
		wtx:   wtx,
		entry: entry,
	}
}

// SaveProtocol is a function that saves the protocol.
func (p *Protocol) SaveProtocol() error {
	protocol, err := util.NewProtocolFromBytes(p.entry.Body)
	if err != nil {
		return err
	}
	switch protocol.Name() {
	case constants.ProtocolCBRC20:
		brc20 := protocol.(*util.CBRC20)
		return p.brc20(brc20)
	}
	return nil
}

// brc20 is a function that saves the brc20c protocol.
func (p *Protocol) brc20(brc20 *util.CBRC20) error {
	switch brc20.Operation {
	case constants.OperationDeploy:
		//list, err := p.wtx.FindProtocol(brc20.Protocol, brc20.Tick, brc20.Operation)
		//if err != nil {
		//	return err
		//}
		//if len(list) > 0 {
		//	log.Log.Warnf("protocol %s, %s, %s already exists", brc20.Protocol, brc20.Tick, brc20.Operation)
		//	return nil
		//}

		if err := p.wtx.SaveProtocol(p.entry.Height, &tables.Protocol{
			InscriptionId: tables.InscriptionId{
				TxId:   p.entry.TxId,
				Offset: p.entry.Offset,
			},
			Index:       p.entry.Index,
			Owner:       p.entry.Owner,
			SequenceNum: p.entry.SequenceNum,
			Protocol:    brc20.Protocol,
			Ticker:      brc20.Tick,
			Operator:    brc20.Operation,
			Max:         gconv.Uint64(brc20.Max),
			Limit:       gconv.Uint64(brc20.Limit),
			Decimals:    gconv.Uint32(brc20.Decimals),
		}); err != nil {
			return err
		}
		//case constants.OperationMint:
		//	inscriptionId := tables.StringToInscriptionId(brc20c.TkId)
		//	deploy, err := p.wtx.GetProtocolByInscriptionId(inscriptionId)
		//	if err != nil {
		//		return err
		//	}
		//	if deploy.Id == 0 {
		//		log.Log.Warnf("tkid: %s not exists", brc20c.TkId)
		//		return nil
		//	}
		//	if deploy.Protocol != constants.OperationDeploy {
		//		log.Log.Warnf("tkid: %s is not deploy operation", brc20c.TkId)
		//		return nil
		//	}
		//	if deploy.Limit < gconv.Uint64(brc20c.Amount) {
		//		log.Log.Warnf("protocol %s, %s, %s amount %s exceeds limit %d", brc20c.Protocol, brc20c.Tick, brc20c.Operation, brc20c.Amount, deploy.Limit)
		//		return nil
		//	}
		//	totalAmount, err := p.wtx.SumProtocolAmount(brc20c.Protocol, brc20c.Tick, brc20c.Operation, brc20c.TkId)
		//	if err != nil {
		//		return err
		//	}
		//	totalAmount += gconv.Uint64(brc20c.Amount)
		//	if totalAmount > deploy.Max {
		//		log.Log.Warnf("protocol %s, %s, %s total amount %d exceeds max %d", brc20c.Protocol, brc20c.Tick, brc20c.Operation, totalAmount, deploy.Max)
		//		return nil
		//	}
		//	mint := &tables.Protocol{
		//		InscriptionId: tables.InscriptionId{
		//			TxId:   p.entry.TxId,
		//			Offset: p.entry.Offset,
		//		},
		//		Index:       p.entry.Index,
		//		SequenceNum: p.entry.SequenceNum,
		//		Protocol:    constants.ProtocolBRC20C,
		//		Ticker:      brc20c.Tick,
		//		Operator:    brc20c.Operation,
		//		TkID:        inscriptionId.String(),
		//		Amount:      gconv.Uint64(brc20c.Amount),
		//		To:          brc20c.To,
		//		Miner:       p.miner,
		//	}
		//	if err := p.wtx.SaveProtocol(mint); err != nil {
		//		return err
		//	}
	}
	return nil
}
