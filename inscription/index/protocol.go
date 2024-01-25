package index

import (
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"github.com/inscription-c/insc/wallet/log"
)

type Protocol struct {
	wtx   *dao.DB
	entry *tables.Inscriptions
	miner string
}

// NewProtocol is a function that returns a new protocol.
func NewProtocol(wtx *dao.DB, entry *tables.Inscriptions, miner string) *Protocol {
	return &Protocol{
		wtx:   wtx,
		entry: entry,
		miner: miner,
	}
}

// SaveProtocol is a function that saves the protocol.
func (p *Protocol) SaveProtocol() error {
	protocol, err := util.NewProtocolFromBytes(p.entry.Body)
	if err != nil {
		return err
	}
	switch protocol.Name() {
	case constants.ProtocolBRC20C:
		brc20c := protocol.(*util.BRC20C)
		return p.brc20c(brc20c)
	}
	return nil
}

// brc20c is a function that saves the brc20c protocol.
func (p *Protocol) brc20c(brc20c *util.BRC20C) error {
	outpoint := p.entry.Outpoint
	switch brc20c.Operation {
	case constants.OperationDeploy:
		list, err := p.wtx.FindProtocol(brc20c.Protocol, brc20c.Tick, brc20c.Operation)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			log.Log.Warnf("protocol %s, %s, %s already exists", brc20c.Protocol, brc20c.Tick, brc20c.Operation)
			return nil
		}

		if err := p.wtx.SaveProtocol(&tables.Protocol{
			Outpoint:    outpoint,
			SequenceNum: p.entry.SequenceNum,
			Protocol:    brc20c.Protocol,
			Ticker:      brc20c.Tick,
			Operator:    brc20c.Operation,
			Max:         gconv.Uint64(brc20c.Max),
			Limit:       gconv.Uint64(brc20c.Limit),
			Decimals:    gconv.Uint32(brc20c.Decimals),
			To:          brc20c.To,
			Miner:       p.miner,
		}); err != nil {
			return err
		}
	case constants.OperationMint:
		inscriptionId := util.StringToInscriptionId(brc20c.TkId)
		deploy, err := p.wtx.GetProtocolByOutpoint(inscriptionId.OutPoint.String())
		if err != nil {
			return err
		}
		if deploy.Id == 0 {
			log.Log.Warnf("tkid: %s not exists", brc20c.TkId)
			return nil
		}
		if deploy.Protocol != constants.OperationDeploy {
			log.Log.Warnf("tkid: %s is not deploy operation", brc20c.TkId)
			return nil
		}
		if deploy.Limit < gconv.Uint64(brc20c.Amount) {
			log.Log.Warnf("protocol %s, %s, %s amount %s exceeds limit %d", brc20c.Protocol, brc20c.Tick, brc20c.Operation, brc20c.Amount, deploy.Limit)
			return nil
		}
		totalAmount, err := p.wtx.CountProtocolAmount(brc20c.Protocol, brc20c.Tick, brc20c.Operation, brc20c.TkId)
		if err != nil {
			return err
		}
		totalAmount += gconv.Uint64(brc20c.Amount)
		if totalAmount > deploy.Max {
			log.Log.Warnf("protocol %s, %s, %s total amount %d exceeds max %d", brc20c.Protocol, brc20c.Tick, brc20c.Operation, totalAmount, deploy.Max)
			return nil
		}
		mint := &tables.Protocol{
			Outpoint:    outpoint,
			SequenceNum: p.entry.SequenceNum,
			Protocol:    constants.ProtocolBRC20C,
			Ticker:      brc20c.Tick,
			Operator:    brc20c.Operation,
			TkID:        inscriptionId.OutPoint.String(),
			Amount:      gconv.Uint64(brc20c.Amount),
			To:          brc20c.To,
			Miner:       p.miner,
		}
		if err := p.wtx.SaveProtocol(mint); err != nil {
			return err
		}
	}
	return nil
}
