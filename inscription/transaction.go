package inscription

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

type Transaction struct {
	Amounts         map[string]uint64
	ChangeAddresses []string
	Recipient       *btcutil.AddressTaproot
}

func (t *Transaction) Build() error {
	//if len(t.ChangeAddresses) != 2 {
	//	return fmt.Errorf("commitTxChangeAddres length must be 2")
	//}
	//for k := range t.ChangeAddresses {
	//	if strings.EqualFold(k, t.Recipient.String()) {
	//		return fmt.Errorf("changeAddres can not be same as commitTxAddress")
	//	}
	//}
	//
	//t.Recipient.ScriptAddress()
	//// TODO dust check
	//
	//delete(t.Utxos, t.Outgoing.String())
	//t.Inputs = append(t.Inputs, t.Outgoing)
	//t.Outputs[t.Recipient.String()] = 0
	return nil
}

func (t *Transaction) SelectOutgoing() {

}

func (t *Transaction) EstimateVBytesWith() uint64 {
	tx := &wire.MsgTx{
		Version: 2,
		TxIn: []*wire.TxIn{{
			PreviousOutPoint: wire.OutPoint{},
			SignatureScript:  []byte{},
			Sequence:         0xFFFFFFFD,
			Witness:          [][]byte{make([]byte, 64)},
		}},
		TxOut: []*wire.TxOut{{
			Value:    0,
			PkScript: t.Recipient.ScriptAddress(),
		}},
	}
	return uint64(tx.SerializeSize())
}
