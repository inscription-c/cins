package index

import (
	"bytes"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gutil"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/model"
)

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

type Envelope struct {
	input   int
	offset  int
	pushNum bool
	stutter bool
	payload *model.Inscription
}

type Envelopes []*Envelope

func ParsedEnvelopFromTransaction(tx *wire.MsgTx) Envelopes {
	return EnvelopeFromRawEnvelope(RawEnvelopeFromTransaction(tx))
}

func EnvelopeFromRawEnvelope(raw RawEnvelopes) Envelopes {
	envelopes := make([]*Envelope, 0)
	for _, r := range raw {
		envelope := fromRawEnvelope(r)
		if envelope != nil {
			envelopes = append(envelopes, envelope)
		}
	}
	return envelopes
}

func fromRawEnvelope(r *RawEnvelope) *Envelope {
	bodyIdx := -1
	for i, v := range r.payload {
		if i%2 == 0 {
			if len(v) == 1 && v[0] == 0 {
				bodyIdx = i
				break
			}
		}
	}
	body := make([]byte, 0)
	if bodyIdx != -1 {
		for i := bodyIdx + 1; i < len(r.payload); i++ {
			body = append(body, r.payload[i]...)
		}
	}

	headIdxEnd := bodyIdx
	if headIdxEnd == -1 {
		headIdxEnd = len(r.payload)
	}

	incompleteField := false
	fields := make(map[TagType][][]byte)
	for i := 0; i < headIdxEnd; i++ {
		if i%2 != 0 {
			continue
		}
		if i+1 < headIdxEnd {
			tag := TagFromBytes(r.payload[i])
			fields[tag] = append(fields[tag], r.payload[i+1])
		} else {
			incompleteField = true
		}
	}

	duplicateField := false
	for _, v := range fields {
		if len(v) > 1 {
			duplicateField = true
			break
		}
	}

	contentEncoding := TagContentEncoding.RemoveField(fields)
	contentType := TagContentType.RemoveField(fields)
	//delegate := TagDelegate.RemoveField(fields)
	metadata := TagMetadata.RemoveField(fields)
	//metaprotocol := TagMetaprotocol.RemoveField(fields)
	//parent := TagParent.RemoveField(fields)
	pointer := TagPointer.RemoveField(fields)
	dstChain := TagDstChain.RemoveField(fields)

	unrecognizedEvenField := false
	for v := range gutil.Keys(fields) {
		if v%2 == 0 {
			unrecognizedEvenField = true
			break
		}
	}

	return &Envelope{
		input:   r.input,
		offset:  r.offset,
		pushNum: r.pushNum,
		stutter: r.stutter,
		payload: &model.Inscription{
			Body:                  body,
			ContentEncoding:       contentEncoding,
			ContentType:           contentType,
			DstChain:              dstChain,
			Metadata:              metadata,
			Pointer:               pointer,
			UnRecognizedEvenField: unrecognizedEvenField,
			DuplicateField:        duplicateField,
			IncompleteField:       incompleteField,
		},
	}

}

type RawEnvelopes []*RawEnvelope

type RawEnvelope struct {
	payload [][]byte
	input   int
	offset  int
	pushNum bool
	stutter bool
}

func RawEnvelopeFromTransaction(tx *wire.MsgTx) RawEnvelopes {
	envelopes := make([]*RawEnvelope, 0)
	for i, input := range tx.TxIn {
		w := &Witness{TxWitness: input.Witness}
		if !w.IsTaprootScript() {
			continue
		}
		stuttered := false
		tokenizer := w.ScriptTokenizer()
		for tokenizer.Next() {
			envelope, stutter := fromInstructions(tokenizer, i, len(envelopes), stuttered)
			if envelope != nil {
				envelopes = append(envelopes, envelope)
			} else {
				stuttered = stutter
			}
		}
	}
	return envelopes
}

type PushBytes []byte

func fromInstructions(
	instructions *txscript.ScriptTokenizer,
	input int,
	offset int,
	stutter bool,
) (*RawEnvelope, bool) {
	if !accept(instructions, txscript.OP_IF) {
		return nil, isPushBytes(instructions.Opcode())
	}
	instructions.Next()
	if !accept(instructions, PushBytes(constants.ProtocolId)) {
		return nil, isPushBytes(instructions.Opcode())
	}

	pushNum := false
	payload := make([][]byte, 0)

	for instructions.Next() {
		switch instructions.Opcode() {
		case txscript.OP_ENDIF:
			return &RawEnvelope{
				input:   input,
				offset:  offset,
				payload: payload,
				pushNum: pushNum,
				stutter: stutter,
			}, false
		case txscript.OP_1NEGATE:
			pushNum = true
			payload = append(payload, []byte{0x81})
		case txscript.OP_0:
			payload = append(payload, []byte{0})
		case txscript.OP_1:
			pushNum = true
			payload = append(payload, []byte{1})
		case txscript.OP_2:
			pushNum = true
			payload = append(payload, []byte{2})
		case txscript.OP_3:
			pushNum = true
			payload = append(payload, []byte{3})
		case txscript.OP_4:
			pushNum = true
			payload = append(payload, []byte{4})
		case txscript.OP_5:
			pushNum = true
			payload = append(payload, []byte{5})
		case txscript.OP_6:
			pushNum = true
			payload = append(payload, []byte{6})
		case txscript.OP_7:
			pushNum = true
			payload = append(payload, []byte{7})
		case txscript.OP_8:
			pushNum = true
			payload = append(payload, []byte{8})
		case txscript.OP_9:
			pushNum = true
			payload = append(payload, []byte{9})
		case txscript.OP_10:
			pushNum = true
			payload = append(payload, []byte{10})
		case txscript.OP_11:
			pushNum = true
			payload = append(payload, []byte{11})
		case txscript.OP_12:
			pushNum = true
			payload = append(payload, []byte{12})
		case txscript.OP_13:
			pushNum = true
			payload = append(payload, []byte{13})
		case txscript.OP_14:
			pushNum = true
			payload = append(payload, []byte{14})
		case txscript.OP_15:
			pushNum = true
			payload = append(payload, []byte{15})
		case txscript.OP_16:
			pushNum = true
			payload = append(payload, []byte{16})
		default:
			if !isPushBytes(instructions.Opcode()) {
				return nil, false
			}
			payload = append(payload, instructions.Data())
		}
	}
	return nil, false
}

func accept(tokenizer *txscript.ScriptTokenizer, instruction interface{}) bool {
	switch instruction.(type) {
	case byte:
		opCode := instruction.(byte)
		return tokenizer.Opcode() == opCode
	case PushBytes:
		pushBytes := instruction.(PushBytes)
		if len(pushBytes) > 0 && bytes.Compare(tokenizer.Data(), pushBytes) != 0 {
			return false
		}
		return true
	default:
		return false
	}
}

func isPushBytes(opcode byte) bool {
	if (opcode >= txscript.OP_DATA_1 && opcode <= txscript.OP_DATA_75) ||
		opcode == txscript.OP_PUSHDATA1 || opcode == txscript.OP_PUSHDATA2 ||
		opcode == txscript.OP_PUSHDATA4 {
		return true
	}
	return false
}
