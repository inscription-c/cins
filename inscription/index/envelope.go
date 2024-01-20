package index

import (
	"bytes"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/model"
)

type Instruction struct {
	opcode      byte
	data        []byte
	isPushBytes bool
}

type Witness struct {
	tokenizer *txscript.ScriptTokenizer
	wire.TxWitness
}

func (w *Witness) IsTaprootScript() bool {
	if len(w.TxWitness) < 2 {
		return false
	}

	posFromLast := 2
	l := len(w.TxWitness)
	lastElement := w.TxWitness[len(w.TxWitness)-1]
	isAnnex := l >= 2 && len(lastElement) > 0 && lastElement[0] == txscript.TaprootAnnexTag
	if isAnnex {
		posFromLast = 3
	}
	if l < posFromLast {
		return false
	}
	tokenizer := txscript.MakeScriptTokenizer(0, w.TxWitness[l-posFromLast])
	w.tokenizer = &tokenizer
	return true
}

func (w *Witness) ScriptTokenizer() *txscript.ScriptTokenizer {
	return w.tokenizer
}

func ParsedEnvelopeFromTransaction(tx *wire.MsgTx) []*Envelope {
	envelopes := make([]*Envelope, 0)
	for i, input := range tx.TxIn {
		w := &Witness{TxWitness: input.Witness}
		if !w.IsTaprootScript() {
			continue
		}
		stuttered := false
		envelope := &Envelope{}
		tokenizer := w.ScriptTokenizer()
		for tokenizer.Next() {
			stuttered = envelope.fromInstructions(tokenizer, i, len(envelopes))
			if !stuttered && envelope.completed {
				envelopes = append(envelopes, envelope)
			}
		}
	}
	return envelopes
}

type Envelope struct {
	input     int
	offset    int
	payload   *model.Inscription
	pushNum   bool
	stutter   bool
	completed bool
}

func (e *Envelope) fromInstructions(
	instructions *txscript.ScriptTokenizer,
	input int,
	offset int,
) bool {
	if !e.accept(instructions, &Instruction{opcode: txscript.OP_IF}) {
		return e.isPushBytes(instructions.Opcode())
	}
	instructions.Next()
	if !e.accept(instructions, &Instruction{isPushBytes: true, data: []byte(constants.ProtocolId)}) {
		return e.isPushBytes(instructions.Opcode())
	}

	latestOpCode := -1
	payload := &model.Inscription{
		Body:     make([]byte, 0),
		Metadata: make([]byte, 0),
		Pointer:  -1,
	}
	for instructions.Next() {
		switch instructions.Opcode() {
		case txscript.OP_ENDIF:
			e.input = input
			e.offset = offset
			e.payload = payload
			e.pushNum = false
			e.stutter = false
			e.completed = true
			return false
		case txscript.OP_1NEGATE:
			e.pushNum = true
			latestOpCode = txscript.OP_1NEGATE
		case txscript.OP_0:
			e.pushNum = true
			latestOpCode = txscript.OP_0
		case txscript.OP_1:
			e.pushNum = true
			latestOpCode = txscript.OP_1
		case txscript.OP_2:
			e.pushNum = true
			latestOpCode = txscript.OP_2
		case txscript.OP_3:
			e.pushNum = true
			latestOpCode = txscript.OP_3
		case txscript.OP_4:
			e.pushNum = true
			latestOpCode = txscript.OP_4
		case txscript.OP_5:
			e.pushNum = true
			latestOpCode = txscript.OP_5
		case txscript.OP_6:
			e.pushNum = true
			latestOpCode = txscript.OP_6
		case txscript.OP_7:
			e.pushNum = true
			latestOpCode = txscript.OP_7
		case txscript.OP_8:
			e.pushNum = true
			latestOpCode = txscript.OP_8
		case txscript.OP_9:
			e.pushNum = true
			latestOpCode = txscript.OP_9
		case txscript.OP_10:
			e.pushNum = true
			latestOpCode = txscript.OP_10
		case txscript.OP_11:
			e.pushNum = true
			latestOpCode = txscript.OP_11
		case txscript.OP_12:
			e.pushNum = true
			latestOpCode = txscript.OP_12
		case txscript.OP_13:
			e.pushNum = true
			latestOpCode = txscript.OP_13
		case txscript.OP_14:
			e.pushNum = true
			latestOpCode = txscript.OP_14
		case txscript.OP_15:
			e.pushNum = true
			latestOpCode = txscript.OP_15
		case txscript.OP_16:
			e.pushNum = true
			latestOpCode = txscript.OP_16
		default:
			if !e.isPushBytes(instructions.Opcode()) {
				return false
			}
			switch latestOpCode {
			case -1:
				switch gconv.Int(string(instructions.Data())) {
				case constants.DstChain:
					latestOpCode = constants.DstChain
				default:
					e.pushNum = true
					return false
				}
			case constants.DstChain:
				if payload.DstChain != "" {
					payload.DuplicateField = true
				}
				payload.DstChain = string(instructions.Data())
			case txscript.OP_0:
				payload.Body = append(payload.Body, instructions.Data()...)
			case txscript.OP_1:
				if payload.ContentType != "" {
					payload.DuplicateField = true
				}
				payload.ContentType = constants.ContentType(instructions.Data())
			case txscript.OP_2:
				if payload.Pointer >= 0 {
					payload.DuplicateField = true
				}
				payload.Pointer = gconv.Int32(string(instructions.Data()))
			case txscript.OP_5:
				payload.Metadata = append(payload.Metadata, instructions.Data()...)
			case txscript.OP_9:
				if payload.ContentEncoding != "" {
					payload.DuplicateField = true
				}
				payload.ContentEncoding = string(instructions.Data())
			}
			latestOpCode = -1
		}
	}
	return false
}

func (e *Envelope) accept(tokenizer *txscript.ScriptTokenizer, instruction *Instruction) bool {
	if instruction.isPushBytes &&
		e.isPushBytes(tokenizer.Opcode()) {
		if len(instruction.data) > 0 && bytes.Compare(tokenizer.Data(), instruction.data) != 0 {
			return false
		}
		return true
	}
	if tokenizer.Opcode() == instruction.opcode {
		return true
	}
	return false
}

func (e *Envelope) isPushBytes(opcode byte) bool {
	if (opcode >= txscript.OP_DATA_1 && opcode <= txscript.OP_DATA_75) ||
		opcode == txscript.OP_PUSHDATA1 || opcode == txscript.OP_PUSHDATA2 ||
		opcode == txscript.OP_PUSHDATA4 {
		return true
	}
	return false
}
