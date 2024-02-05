package index

import (
	"bytes"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gutil"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
)

// Witness is a struct that represents a witness in a transaction.
// It contains two fields:
// - tokenizer: a pointer to a ScriptTokenizer from the txscript package. This is used to tokenize the script in the witness.
// - TxWitness: an embedded field from the wire package. This represents the witness in a transaction.
type Witness struct {
	tokenizer      *txscript.ScriptTokenizer // The script tokenizer for the witness
	wire.TxWitness                           // The witness in a transaction
}

// IsTaprootScript checks if the witness is a Taproot script.
// It first checks if the length of the witness is less than 2, if so it returns false.
// It then checks if the last element of the witness is an annex, if so it sets the position from last to 3.
// If the length of the witness is less than the position from last, it returns false.
// It then creates a script tokenizer from the witness at the position from last and sets the witness's tokenizer to this new tokenizer.
// It finally returns true, indicating that the witness is a Taproot script.
func (w *Witness) IsTaprootScript() bool {
	// Check if the length of the witness is less than 2
	if len(w.TxWitness) < 2 {
		return false
	}

	// Initialize the position from last to 2
	posFromLast := 2
	// Get the length of the witness
	l := len(w.TxWitness)
	// Get the last element of the witness
	lastElement := w.TxWitness[l-1]
	// Check if the last element is an annex
	isAnnex := l >= 2 && len(lastElement) > 0 && lastElement[0] == txscript.TaprootAnnexTag
	// If the last element is an annex, set the position from last to 3
	if isAnnex {
		posFromLast = 3
	}
	// If the length of the witness is less than the position from last, return false
	if l < posFromLast {
		return false
	}
	// Create a script tokenizer from the witness at the position from last
	tokenizer := txscript.MakeScriptTokenizer(0, w.TxWitness[l-posFromLast])
	// Set the witness's tokenizer to the new tokenizer
	w.tokenizer = &tokenizer
	// Return true, indicating that the witness is a Taproot script
	return true
}

// ScriptTokenizer returns the script tokenizer of the witness.
func (w *Witness) ScriptTokenizer() *txscript.ScriptTokenizer {
	return w.tokenizer
}

// Envelope represents an envelope in a transaction.
// It is a struct that contains the following fields:
// - input: an integer that represents the index of the input in the transaction.
// - offset: an integer that represents the offset of the envelope in the transaction.
// - pushNum: a boolean that indicates whether the envelope is a push number envelope.
// - stutter: a boolean that indicates whether the envelope is stuttered.
// - payload: a pointer to an Inscription struct that represents the payload of the envelope.
type Envelope struct {
	index   uint32
	owner   string
	offset  uint32
	pushNum bool
	stutter bool
	payload *model.Inscription
}

// Envelopes is a slice of pointers to Envelope.
type Envelopes []*Envelope

// ParsedEnvelopFromTransaction parses envelopes from a transaction.
func ParsedEnvelopFromTransaction(tx *wire.MsgTx) Envelopes {
	return EnvelopeFromRawEnvelope(RawEnvelopeFromTransaction(tx))
}

// EnvelopeFromRawEnvelope is a function that creates envelopes from raw envelopes.
// It takes a slice of pointers to RawEnvelope as a parameter and returns a slice of pointers to Envelope.
// The function iterates over each raw envelope in the slice. For each raw envelope, it creates an envelope.
// If the envelope is not nil, it appends the envelope to the slice of envelopes.
// After iterating over all raw envelopes, it returns the slice of envelopes.
func EnvelopeFromRawEnvelope(raw RawEnvelopes) Envelopes {
	envelopes := make([]*Envelope, 0)
	for _, r := range raw {
		// Create an envelope from the raw envelope
		envelope := fromRawEnvelope(r)
		if envelope != nil {
			envelopes = append(envelopes, envelope)
		}
	}
	return envelopes
}

// fromRawEnvelope is a function that creates an envelope from a raw envelope.
// It takes a pointer to a RawEnvelope as a parameter and returns a pointer to an Envelope.
// The function first finds the index of the body in the payload of the raw envelope.
// It then creates the body by appending all payloads after the body index.
// It also creates a map of fields from the payloads before the body index.
// It checks for incomplete and duplicate fields in the map.
// It then removes recognized fields from the map and assigns them to their respective variables.
// It checks for unrecognized even fields in the remaining map.
// Finally, it creates an Envelope with the input, offset, pushNum, stutter, and payload from the raw envelope and the fields from the map.
// The payload of the Envelope is an Inscription that contains the body, contentEncoding, contentType, dstChain, metadata, pointer, and flags for unrecognized even field, duplicate field, and incomplete field.
func fromRawEnvelope(r *RawEnvelope) *Envelope {
	// Find the index of the body in the payload of the raw envelope
	bodyIdx := -1
	for i, v := range r.payload {
		if i%2 == 0 {
			if len(v) == 1 && v[0] == 0 {
				bodyIdx = i
				break
			}
		}
	}
	// Create the body by appending all payloads after the body index
	body := make([]byte, 0)
	if bodyIdx != -1 {
		for i := bodyIdx + 1; i < len(r.payload); i++ {
			body = append(body, r.payload[i]...)
		}
	}

	// Create a map of fields from the payloads before the body index
	headIdxEnd := bodyIdx
	if headIdxEnd == -1 {
		headIdxEnd = len(r.payload)
	}

	// Initialize a flag to indicate if there is an incomplete field
	incompleteField := false
	// Initialize a map to store the fields. The key is a TagType and the value is a 2D byte slice.
	fields := make(map[TagType][][]byte)
	// Iterate over the payloads before the body index
	for i := 0; i < headIdxEnd; i++ {
		// If the index is odd, skip the current iteration
		if i%2 != 0 {
			continue
		}
		// If the index is even and there is a next payload, add the payload to the fields map
		if i+1 < headIdxEnd {
			// Convert the payload to a TagType
			tag := TagFromBytes(r.payload[i])
			// Append the next payload to the current tag in the fields map
			fields[tag] = append(fields[tag], r.payload[i+1])
		} else {
			// If the index is even and there is no next payload, set the incomplete field flag to true
			incompleteField = true
		}
	}

	// Check for duplicate fields in the map
	duplicateField := false
	for _, v := range fields {
		if len(v) > 1 {
			duplicateField = true
			break
		}
	}

	// Remove recognized fields from the map and assign them to their respective variables
	contentEncoding := TagContentEncoding.RemoveField(fields)
	contentType := TagContentType.RemoveField(fields)
	metadata := TagMetadata.RemoveField(fields)
	pointer := TagPointer.RemoveField(fields)
	unlockConditionData := TagUnlockCondition.RemoveField(fields)

	// Check for unrecognized even fields in the remaining map
	unrecognizedEvenField := false
	for v := range gutil.Keys(fields) {
		if v%2 == 0 {
			unrecognizedEvenField = true
			break
		}
	}

	inscription := &model.Inscription{
		Body:                  body,
		ContentEncoding:       contentEncoding,
		ContentType:           constants.ContentType(contentType),
		Metadata:              metadata,
		Pointer:               pointer,
		UnRecognizedEvenField: unrecognizedEvenField,
		DuplicateField:        duplicateField,
		IncompleteField:       incompleteField,
	}

	// Create an UnlockCondition from the unlock condition data
	unlockCondition, err := tables.UnlockConditionFromBytes(unlockConditionData)
	if err != nil {
		unrecognizedEvenField = true
	} else {
		inscription.UnlockCondition = *unlockCondition
	}

	return &Envelope{
		owner:   r.owner,
		index:   r.index,
		offset:  r.offset,
		pushNum: r.pushNum,
		stutter: r.stutter,
		payload: inscription,
	}
}

// RawEnvelopes is a slice of pointers to RawEnvelope.
type RawEnvelopes []*RawEnvelope

// RawEnvelope represents a raw envelope in a transaction.
// It is a struct that contains the following fields:
// - payload: a 2D byte slice that represents the payload of the envelope.
// - input: an integer that represents the index of the input in the transaction.
// - offset: an integer that represents the offset of the envelope in the transaction.
// - pushNum: a boolean that indicates whether the envelope is a push number envelope.
// - stutter: a boolean that indicates whether the envelope is stuttered.
type RawEnvelope struct {
	payload [][]byte
	owner   string
	index   uint32
	offset  uint32
	pushNum bool
	stutter bool
}

// RawEnvelopeFromTransaction is a function that creates raw envelopes from a transaction.
// It takes a transaction as a parameter and returns a slice of pointers to RawEnvelope.
// The function iterates over each input in the transaction. For each input, it creates a Witness and checks if it is a Taproot script.
// If the Witness is not a Taproot script, it continues to the next input.
// If the Witness is a Taproot script, it creates a script tokenizer and iterates over the instructions in the script.
// For each instruction, it creates a raw envelope and checks if the envelope is not nil.
// If the envelope is not nil, it appends the envelope to the slice of envelopes.
// If the envelope is nil, it sets the stuttered flag to the value of stutter.
// After iterating over all inputs and instructions, it returns the slice of envelopes.
func RawEnvelopeFromTransaction(tx *wire.MsgTx) RawEnvelopes {
	envelopes := make([]*RawEnvelope, 0)
	for index, input := range tx.TxIn {
		w := &Witness{TxWitness: input.Witness}
		if !w.IsTaprootScript() {
			continue
		}

		stuttered := false
		var owner string
		if index == 0 {
			_, address, _, err := txscript.ExtractPkScriptAddrs(tx.TxOut[index].PkScript, util.ActiveNet.Params)
			if err != nil {
				continue
			}
			if len(address) > 0 {
				owner = address[0].String()
			}
		}
		tokenizer := w.ScriptTokenizer()
		for tokenizer.Next() {
			// Create a raw envelope from the instructions
			envelope, stutter := fromInstructions(tokenizer, uint32(index), uint32(len(envelopes)), stuttered)
			// Check if the envelope is not nil
			if envelope != nil {
				envelope.owner = owner
				envelopes = append(envelopes, envelope)
			} else {
				// If the envelope is nil, set the stuttered flag to the value of stutter
				stuttered = stutter
			}
		}
	}
	return envelopes
}

// PushBytes represents a slice of bytes to be pushed.
type PushBytes []byte

// fromInstructions creates a raw envelope from instructions.
// It takes a script tokenizer, an input index, an offset, and a stutter flag as parameters.
// It returns a pointer to a RawEnvelope and a boolean value.
// The boolean value indicates whether the opcode is a push bytes opcode.
func fromInstructions(
	instructions *txscript.ScriptTokenizer,
	index uint32,
	offset uint32,
	stutter bool,
) (*RawEnvelope, bool) {
	// If the opcode does not match OP_IF, return nil and whether the opcode is a push bytes opcode
	if !accept(instructions, txscript.OP_IF) {
		return nil, isPushBytes(instructions.Opcode())
	}
	instructions.Next()
	// If the opcode does not match the protocol ID, return nil and whether the opcode is a push bytes opcode
	if !accept(instructions, PushBytes(constants.ProtocolId)) {
		return nil, isPushBytes(instructions.Opcode())
	}

	pushNum := false
	payload := make([][]byte, 0)

	for instructions.Next() {
		switch instructions.Opcode() {
		case txscript.OP_ENDIF:
			return &RawEnvelope{
				index:   index,
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
			pushNum = false
			payload = append(payload, instructions.Data())
		}
	}
	return nil, false
}

// accept checks if the tokenizers opcode matches the instruction.
func accept(tokenizer *txscript.ScriptTokenizer, instruction interface{}) bool {
	switch instruction.(type) {
	case int:
		tokenOpCode := tokenizer.Opcode()
		opCode := byte(instruction.(int))
		return tokenOpCode == opCode
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

// isPushBytes checks if the opcode is a push bytes opcode.
func isPushBytes(opcode byte) bool {
	if (opcode >= txscript.OP_DATA_1 && opcode <= txscript.OP_DATA_75) ||
		opcode == txscript.OP_PUSHDATA1 || opcode == txscript.OP_PUSHDATA2 ||
		opcode == txscript.OP_PUSHDATA4 {
		return true
	}
	return false
}
