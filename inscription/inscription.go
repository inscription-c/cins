package inscription

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	txscript2 "github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/waddrmgr"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/go-playground/validator/v10"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/indexer"
	"github.com/inscription-c/insc/internal/util"
	"github.com/inscription-c/insc/internal/util/txscript"
	"github.com/shopspring/decimal"
	"github.com/ugorji/go/codec"
	"io"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

// Package validator is imported to perform validation on the options' struct.
var validate = validator.New(validator.WithRequiredStructEnabled())

// Inscription is a struct that represents an inscription in the blockchain.
// It contains the body of the inscription, the header, the body protocol,
// the options for the inscription, the transaction outputs, the commit and reveal transactions,
// the private and internal keys, the reveal script, the tap leaf node, the control block,
// the script builder, and the reveal transaction address.
type Inscription struct {
	// body is the body of the inscription.
	body []byte

	// Header is the header of the inscription.
	Header Header

	// Body is the protocol of the body of the inscription.
	Body util.Protocol

	// options are the options for the inscription.
	options *options

	// vout is the index of the output in the transaction.
	vout int

	// feeRate is the fee rate for the transaction.
	feeRate int64

	// revealFee is the fee for the reveal transaction.
	revealFee int64

	totalFee int64

	// utxo is the unspent transaction outputs for the wallet.
	utxo []btcjson.ListUnspentResult

	// commitTx is the commit transaction of the inscription.
	// revealTx is the reveal transaction of the inscription.
	commitTx, revealTx *wire.MsgTx

	// priKey is the private key for the transaction.
	priKey *btcec.PrivateKey

	// internalKey is the internal key for the transaction.
	internalKey *btcec.PublicKey

	// revealScript is the reveal script for the transaction.
	revealScript []byte

	// tapLeafNode is the tap leaf node for the transaction.
	tapLeafNode txscript.TapLeaf

	// controlBlock is the control block for the transaction.
	controlBlock *txscript2.ControlBlock

	// scriptBuilder is the script builder for the transaction.
	scriptBuilder *txscript.ScriptBuilder

	// revealTxAddress is the address for the reveal transaction.
	revealTxAddress *btcutil.AddressTaproot
}

// Header is a struct that represents the header of an inscription.
// It contains the destination chain, the content type, the content encoding,
// the pointer, and the metadata.
type Header struct {
	// CInsDescription is the destination chain for the inscription.
	CInsDescription *tables.CInsDescription `json:"c_ins_description"`

	// ContentType is the content type of the inscription.
	ContentType constants.ContentType `json:"content_type"`

	// ContentEncoding is the encoding of the content of the inscription.
	ContentEncoding string `json:"content_encoding"`

	// Pointer is the pointer to the content of the inscription.
	Pointer string `json:"pointer"`

	// Metadata is the metadata of the inscription.
	Metadata *util.Reader `json:"metadata"`
}

// options is a struct that represents the options for an inscription.
// It contains the wallet client, the postage, the wallet password, the destination chain,
// the CBOR metadata, and the JSON metadata.
type options struct {
	// walletClient is the client for the wallet.
	walletClient *rpcclient.Client `validate:"required"`

	// postage is the postage for the inscription.
	postage uint64 `validate:"required"`

	// walletPass is the password for the wallet.
	walletPass string `validate:"required"`

	// cInsDescription is the destination chain for the inscription.
	cInsDescription *tables.CInsDescription

	// cborMetadata is the CBOR metadata for the inscription.
	cborMetadata string

	// jsonMetadata is the JSON metadata for the inscription.
	jsonMetadata string

	indexer *indexer.Indexer
}

// Option is a function type that takes a pointer to an options' struct.
// It is used to set the options for an inscription.
type Option func(*options)

// WithPostage is a function that sets the postage option for an Inscription.
// It takes an uint64 value representing the postage and returns a function that
// sets the postage in the options of an Inscription.
func WithPostage(postage uint64) func(*options) {
	return func(options *options) {
		options.postage = postage
	}
}

// WithWalletPass is a function that sets the wallet password option for an Inscription.
// It takes a string representing the wallet password and returns a function that sets the
// wallet password in the options of an Inscription.
func WithWalletPass(pass string) func(*options) {
	return func(options *options) {
		options.walletPass = pass
	}
}

// WithCborMetadata is a function that sets the CBOR metadata option for an Inscription.
// It takes a string representing the CBOR metadata and returns a function that sets the
// CBOR metadata in the options of an Inscription.
func WithCborMetadata(cborMetadata string) func(*options) {
	return func(options *options) {
		options.cborMetadata = cborMetadata
	}
}

// WithJsonMetadata is a function that sets the JSON metadata option for an Inscription.
// It takes a string representing the JSON metadata and returns a function that sets the
// JSON metadata in the options of an Inscription.
func WithJsonMetadata(jsonMetadata string) func(*options) {
	return func(options *options) {
		options.jsonMetadata = jsonMetadata
	}
}

// WithCInsDescription is a function that sets the destination chain option for an Inscription.
// It takes a string representing the destination chain and returns a function that sets
// the destination chain in the options of an Inscription.
func WithCInsDescription(cInsDescription *tables.CInsDescription) func(*options) {
	return func(options *options) {
		options.cInsDescription = cInsDescription
	}
}

// WithWalletClient is a function that sets the wallet client option for an Inscription.
// It takes a pointer to a rpcclient.Client representing the wallet client and returns a
// function that sets the wallet client in the options of an Inscription.
func WithWalletClient(cli *rpcclient.Client) func(*options) {
	return func(options *options) {
		options.walletClient = cli
	}
}

func WithIndexer(indexer *indexer.Indexer) func(*options) {
	return func(options *options) {
		options.indexer = indexer
	}
}

// NewFromPath is a function that creates a new Inscription from a given path.
// It takes a string representing the path and a variadic number of Option functions
// to set the options for the Inscription. It validates the options, sets the options
// in the Inscription, parses the CBOR and JSON metadata, sets the content type and
// content encoding, and sets the body of the Inscription. It returns a pointer to the
// created Inscription and any error that occurred during the process.
func NewFromPath(path string, inputOpts ...Option) (*Inscription, error) {
	// Create a new options struct and apply all the provided options to it
	opts := &options{}
	for _, option := range inputOpts {
		option(opts)
	}

	// Create maps to hold the options and their validation rules
	optsMap := make(map[string]interface{})
	rules := make(map[string]interface{})
	// Get the value and type of the options struct
	v := reflect.ValueOf(opts).Elem()
	t := reflect.TypeOf(opts).Elem()
	// Iterate over the fields of the options struct
	for i := 0; i < t.NumField(); i++ {
		// Get the value of the field
		fv := v.Field(i)
		// Get the name of the field
		fieldName := t.Field(i).Name
		// Get the value of the field as an interface
		value := reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
		// Add the field to the options map
		optsMap[fieldName] = value.Interface()
		// Get the validation tag of the field
		tag := t.Field(i).Tag.Get("validate")
		// If the field has a validation tag, add it to the rules map
		if tag != "" {
			rules[fieldName] = tag
		}
	}

	// Validate the options using the validation rules
	if errs := validate.ValidateMap(optsMap, rules); len(errs) > 0 {
		return nil, errors.New(fmt.Sprint(errs))
	}

	// Create a new Inscription with the provided options
	inscription := &Inscription{
		Header: Header{
			CInsDescription: opts.cInsDescription,
			Metadata:        &util.Reader{},
		},
		options: opts,
	}

	// Parse the CBOR metadata if provided
	if err := inscription.parseCborMetadata(opts.cborMetadata); err != nil {
		return nil, err
	}
	// Parse the JSON metadata if provided
	if err := inscription.parseJsonMetadata(opts.jsonMetadata); err != nil {
		return nil, err
	}

	// Determine the content type of the file at the provided path
	media, err := util.ContentTypeForPath(path)
	if err != nil {
		return nil, err
	}
	// Set the content type of the Inscription
	inscription.Header.ContentType = media.ContentType

	// Read the file at the provided path
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Initialize the content encoding as an empty string
	contentEncoding := ""

	// If compression is enabled, compress the body
	if compress {
		buf := bytes.NewBufferString("")
		bw := brotli.NewWriterOptions(buf, brotli.WriterOptions{Quality: 11, LGWin: 24})
		if _, err := bw.Write(body); err != nil {
			return nil, err
		}
		if err := bw.Close(); err != nil {
			return nil, err
		}
		decompressed := make([]byte, 0)
		if _, err := brotli.NewReader(buf).Read(decompressed); err != nil {
			return nil, err
		}
		if bytes.Compare(body, decompressed) != 0 {
			return nil, errors.New("decompression round trip failed")
		}

		// If the compressed body is smaller than the original body, use the compressed body and set the content encoding to "br"
		if len(buf.Bytes()) < len(body) {
			body = buf.Bytes()
			contentEncoding = "br"
		}
	}
	// Set the body and content encoding of the Inscription
	inscription.body = body
	inscription.Header.ContentEncoding = contentEncoding

	// Initialize the body protocol of the Inscription
	var incBody util.Protocol
	if cbrc20 {
		incBody = &util.CBRC20{}
	} else {
		incBody = &util.DefaultProtocol{}
	}
	incBody.Reset(body)

	// Set the body protocol of the Inscription
	inscription.Body = incBody
	return inscription, nil
}

// Wallet is a method of the Inscription struct. It returns the wallet client of the Inscription.
func (i *Inscription) Wallet() *rpcclient.Client {
	return i.options.walletClient
}

// parseCborMetadata is a method of the Inscription struct. It is responsible
// for parsing the CBOR metadata of the Inscription. It reads the CBOR metadata
// from a file, decodes the CBOR metadata, and sets the decoded metadata in the Inscription.
// It returns an error if there is an error in any of the steps.
func (i *Inscription) parseCborMetadata(cborMetadata string) error {
	if cborMetadata == "" {
		return nil
	}
	data, err := os.ReadFile(cborMetadata)
	if err != nil {
		return err
	}
	var metadata interface{}
	if err := codec.NewDecoderBytes(data, &codec.CborHandle{}).Decode(&metadata); err != nil {
		return err
	}
	i.Header.Metadata.Reset(data)
	return nil
}

// parseJsonMetadata is a method of the Inscription struct. It is responsible
// for parsing the JSON metadata of the Inscription. It reads the JSON metadata
// from a file, unmarshals the JSON metadata, and sets the unmarshaled metadata
// in the Inscription. It returns an error if there is an error in any of the steps.
func (i *Inscription) parseJsonMetadata(jsonMetadata string) error {
	if jsonMetadata == "" {
		return nil
	}
	data, err := os.ReadFile(jsonMetadata)
	if err != nil {
		return err
	}
	var metadata interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}
	i.Header.Metadata.Reset(data)
	return nil
}

// CommitTxId is a method of the Inscription struct. It returns the transaction
// ID of the commit transaction of the Inscription.
func (i *Inscription) CommitTxId() string {
	return i.commitTx.TxHash().String()
}

// RevealTxId is a method of the Inscription struct. It returns the transaction
// ID of the reveal transaction of the Inscription.
func (i *Inscription) RevealTxId() string {
	return i.revealTx.TxHash().String()
}

// rawTx is a method of the Inscription struct. It is responsible for serializing
// a given transaction. It takes a pointer to a wire.MsgTx representing the transaction
// and a variadic number of booleans representing whether to serialize without witness.
// It serializes the transaction with or without witness based on the boolean and returns
// the serialized transaction.
func (i *Inscription) rawTx(tx *wire.MsgTx, noWitness ...bool) string {
	buf := bytes.NewBufferString("")
	if len(noWitness) > 0 && noWitness[0] {
		if err := tx.SerializeNoWitness(buf); err != nil {
			return ""
		}
	} else {
		if err := tx.Serialize(buf); err != nil {
			return ""
		}
	}
	return hex.EncodeToString(buf.Bytes())
}

// CreateInscriptionTx is a method of the Inscription struct. It is responsible for
// creating the inscription transaction of the Inscription. It estimates the fee rate,
// generates a temporary private key, builds the reveal transaction, and builds the
// commit transaction. It returns an error if there is an error in any of the steps.
func (i *Inscription) CreateInscriptionTx() error {
	// get fee rate
	feeRate, err := i.Wallet().EstimateFee(10)
	if err != nil {
		return err
	}
	i.feeRate = int64(index.AmountToSat(feeRate))

	// gen temporary priKey
	priKey, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}
	i.priKey = priKey

	// build reveal tx
	if err := i.BuildRevealTx(); err != nil {
		return err
	}

	// build commit tx
	if err := i.BuildCommitTx(); err != nil {
		return err
	}
	return nil
}

// BuildCommitTx is a method of the Inscription struct. It is responsible
// for building the commit transaction of the Inscription. It initializes
// the total input and output amounts, creates the transaction inputs and
// outputs, calculates the change, creates the change output, and clears
// the input scripts for the transaction. It returns an error if there is
// an error in any of the steps.
func (i *Inscription) BuildCommitTx() error {
	var inTotal, outTotal int64
	commitTx := wire.NewMsgTx(2)
	i.commitTx = commitTx

	// input begin
	for _, v := range i.utxo {
		script, err := util.AddressScript(v.Address, util.ActiveNet.Params)
		if err != nil {
			return err
		}
		hash, err := chainhash.NewHashFromStr(v.TxID)
		if err != nil {
			return err
		}
		txIn := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: v.Vout,
		}, script, nil)
		commitTx.AddTxIn(txIn)
		inTotal += decimal.NewFromFloat(v.Amount).Mul(decimal.NewFromInt(int64(constants.OneBtc))).IntPart()
	}
	// input end

	// output begin
	recipientScript, err := util.AddressScript(i.revealTxAddress.String(), util.ActiveNet.Params)
	if err != nil {
		return err
	}
	commitTx.AddTxOut(wire.NewTxOut(int64(postage)+i.revealFee, recipientScript))
	outTotal += int64(postage) + i.revealFee
	// output end

	// change calculate
	change := inTotal - outTotal - calculateTxFee(commitTx, i.feeRate)
	if change < 0 {
		return InsufficientBalanceError
	}

	// change output
	commitTxChangeAddr, err := i.Wallet().GetRawChangeAddressType(constants.DefaultWalletName, constants.AddressTypeP2shSegWit)
	if err != nil {
		return err
	}
	changeScript, err := util.AddressScript(commitTxChangeAddr.String(), util.ActiveNet.Params)
	if err != nil {
		return err
	}
	commitTx.AddTxOut(wire.NewTxOut(change, changeScript))
	fee := calculateTxFee(commitTx, i.feeRate)
	i.totalFee += fee
	change = inTotal - outTotal - fee
	commitTx.TxOut[len(commitTx.TxOut)-1].Value = change
	if change < constants.DustLimit {
		commitTx.TxOut = commitTx.TxOut[:len(commitTx.TxOut)-1]
	}

	//delete input script
	for _, v := range commitTx.TxIn {
		v.SignatureScript = nil
	}
	return nil
}

// BuildRevealTx is a method of the Inscription struct. It is responsible
// for building the reveal transaction of the Inscription. It generates a
// temporary key, builds the reveal script, creates the reveal transaction,
// and clears the input scripts for the transaction. It returns an error if
// there is an error in any of the steps.
func (i *Inscription) BuildRevealTx() error {
	// Generate a temporary key
	internalKey := i.priKey.PubKey()
	i.internalKey = internalKey

	// Start building the reveal script
	// Append check sign op to the script
	i.scriptBuilder = txscript.NewScriptBuilder()
	i.scriptBuilder.AddData(schnorr.SerializePubKey(internalKey))
	i.scriptBuilder.AddOp(txscript.OP_CHECKSIG)
	// Append the inscription content to the script builder
	if err := i.AppendInscriptionContentToBuilder(); err != nil {
		return err
	}
	revealScript, err := i.scriptBuilder.Script()
	if err != nil {
		return err
	}
	i.revealScript = revealScript

	// Generate the script address
	if err := i.RevealScriptAddress(); err != nil {
		return err
	}
	log.Log.Info("taprootAddress", i.revealTxAddress.String())

	// Create the witness for the transaction
	revealTxWitness := make([][]byte, 0)
	revealTxWitness = append(revealTxWitness, make([]byte, 64))
	revealTxWitness = append(revealTxWitness, revealScript)
	controlBlockBytes, err := i.controlBlock.ToBytes()
	if err != nil {
		return err
	}
	revealTxWitness = append(revealTxWitness, controlBlockBytes)
	taprootScript, err := txscript.PayToAddrScript(i.revealTxAddress)
	if err != nil {
		return err
	}

	// Create the transaction input
	revealTxIn := &wire.TxIn{
		SignatureScript: taprootScript,
		Witness:         revealTxWitness,
		Sequence:        0xFFFFFFFD,
	}

	// Create the transaction output
	destAddrScript, err := util.AddressScript(strings.TrimSpace(destination), util.ActiveNet.Params)
	if err != nil {
		return err
	}
	revealTxOutput := wire.NewTxOut(int64(postage), destAddrScript)

	// Create the reveal transaction
	revealTx := wire.NewMsgTx(2)
	i.revealTx = revealTx
	revealTx.AddTxIn(revealTxIn)
	revealTx.AddTxOut(revealTxOutput)
	i.revealFee = calculateTxFee(revealTx, i.feeRate)
	i.totalFee += i.revealFee

	// Clear the input scripts for the transaction
	revealTxIn.SignatureScript = nil
	return nil
}

// SignCommitTx is a method of the Inscription struct. It is responsible for
// signing the commit transaction of the Inscription. It unlocks the wallet,
// fetches the private keys for the transaction inputs, calculates the signature
// hashes, and signs the transaction inputs. It returns an error if there is an
// error in any of the steps.
func (i *Inscription) SignCommitTx() error {
	// This block of code is part of the signCommitTx method of the Inscription struct.
	// It is responsible for signing the commit transaction of the Inscription.

	// First, it unlocks the wallet using the wallet passphrase.
	if err := i.Wallet().WalletPassphrase(walletPass, 5); err != nil {
		return err
	}

	// It creates a map to hold the private keys for the transaction inputs.
	priKeyMap := make(map[string]*btcutil.WIF)

	// It creates a new MultiPrevOutFetcher to fetch previous outputs.
	feature := txscript.NewMultiPrevOutFetcher(nil)

	// It iterates over the unspent transaction outputs (UTXOs) of the Inscription.
	for j := 0; j < len(i.utxo); j++ {
		utxo := i.utxo[j]

		// It decodes the address of the UTXO.
		address, err := btcutil.DecodeAddress(utxo.Address, util.ActiveNet.Params)
		if err != nil {
			return err
		}

		// It dumps the private key for the address.
		wif, err := i.Wallet().DumpPrivKey(address)
		if err != nil {
			return err
		}

		// It creates a new OutPoint for the UTXO and adds the private key to the map.
		outpoint := model.NewOutPoint(utxo.TxID, utxo.Vout)
		priKeyMap[outpoint.String()] = wif

		// It converts the OutPoint to a wire.OutPoint.
		outpointObj, err := outpoint.WireOutpoint()
		if err != nil {
			return err
		}

		// It decodes the script public key of the UTXO.
		pkScript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return err
		}

		// It converts the amount of the UTXO to satoshis.
		value := int64(index.Amount(utxo.Amount).Sat())

		// It adds the previous output to the MultiPrevOutFetcher.
		feature.AddPrevOut(*outpointObj, wire.NewTxOut(value, pkScript))
	}

	// It creates new transaction signature hashes using the commit transaction and the MultiPrevOutFetcher.
	sigHashes := txscript.NewTxSigHashes(i.commitTx, feature)

	// It iterates over the UTXOs of the Inscription again.
	for j := 0; j < len(i.utxo); j++ {
		utxo := i.utxo[j]

		// It creates a new OutPoint for the UTXO and retrieves the private key from the map.
		outpoint := model.NewOutPoint(utxo.TxID, utxo.Vout)
		wif := priKeyMap[outpoint.String()]

		// It converts the address of the UTXO to a script.
		pkScript, err := util.AddressScript(utxo.Address, util.ActiveNet.Params)
		if err != nil {
			return err
		}

		// It converts the amount of the UTXO to satoshis.
		value := int64(index.Amount(utxo.Amount).Sat())

		// It creates a witness signature for the transaction input.
		witness, err := txscript.WitnessSignature(i.commitTx, sigHashes, j, value, pkScript, txscript.SigHashAll, wif.PrivKey, wif.CompressPubKey)
		if err != nil {
			return err
		}

		// It sets the witness of the transaction input.
		i.commitTx.TxIn[j].Witness = witness
	}

	// It returns nil if there were no errors during the process.
	return nil
}

// SignRevealTx is a method of the Inscription struct. It is responsible
// for signing the reveal transaction of the Inscription. It sets the previous
// outpoint of the reveal transaction input, calculates the signature hash, and
// signs the reveal transaction input. It returns an error if there is an error in any of the steps.
func (i *Inscription) SignRevealTx() error {
	// This block of code is part of the signRevealTx method of the Inscription struct.
	// It is responsible for signing the reveal transaction of the Inscription.

	// First, it gets the hash of the commit transaction and sets it as the previous outpoint of the reveal transaction input.
	commitHash := i.commitTx.TxHash()
	i.revealTx.TxIn[i.vout].PreviousOutPoint = *wire.NewOutPoint(&commitHash, uint32(i.vout))

	// It creates a new MultiPrevOutFetcher to fetch previous outputs.
	prevFetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
		i.revealTx.TxIn[i.vout].PreviousOutPoint: {
			Value:    i.commitTx.TxOut[i.vout].Value,
			PkScript: i.commitTx.TxOut[i.vout].PkScript,
		},
	})

	// It creates new transaction signature hashes using the reveal transaction and the MultiPrevOutFetcher.
	sigHashes := txscript.NewTxSigHashes(i.revealTx, prevFetcher)

	// It calculates the signature hash for the reveal transaction.
	signHash, err := txscript.CalcTapscriptSignaturehash(sigHashes, txscript.SigHashDefault, i.revealTx, 0, prevFetcher, i.tapLeafNode)
	if err != nil {
		return err
	}

	// It signs the signature hash using the private key.
	signature, err := schnorr.Sign(i.priKey, signHash)
	if err != nil {
		return err
	}

	// It serializes the signature and sets it as the witness of the reveal transaction input.
	sig := signature.Serialize()
	i.revealTx.TxIn[i.vout].Witness[0] = sig
	return nil
}

// AppendInscriptionContentToBuilder is a method of the Inscription struct. It is
// responsible for appending the reveal script to the script builder. It adds the
// protocol ID, content type, metadata, content encoding, and body to the script builder.
// It returns an error if there is an error in any of the steps.
func (i *Inscription) AppendInscriptionContentToBuilder() error {
	// This block of code is part of the appendRevealScriptToBuilder method of the Inscription struct.
	// It is responsible for appending the reveal script to the script builder.

	// Start building the reveal script
	// Add the initial operations and the protocol ID to the script builder
	i.scriptBuilder.
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte(constants.ProtocolId)).
		AddData([]byte(constants.CInsDescription)).
		AddData(i.Header.CInsDescription.Data()).
		AddOp(txscript.OP_1).
		AddData(i.Header.ContentType.Bytes())

	// If metadata exists, add it to the script builder
	// The metadata is divided into chunks of 520 bytes and each chunk is added to the script builder
	if i.Header.Metadata.Len() > 0 {
		for {
			data, err := i.Header.Metadata.Chunks(520)
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			i.scriptBuilder.AddOp(txscript.OP_5)
			i.scriptBuilder.AddData(data)
		}
	}

	// If content encoding exists, add it to the script builder
	if i.Header.ContentEncoding != "" {
		i.scriptBuilder.AddOp(txscript.OP_9)
		i.scriptBuilder.AddData([]byte(i.Header.ContentEncoding))
	}

	// If body exists, add it to the script builder
	// The body is divided into chunks of 520 bytes and each chunk is added to the script builder
	if i.Body.Len() > 0 {
		for {
			body, err := i.Body.Chunks(520)
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			i.scriptBuilder.AddOp(txscript.OP_0)
			i.scriptBuilder.AddData(body)
		}
	}

	// Add the end if operation to the script builder
	i.scriptBuilder.AddOp(txscript.OP_ENDIF)
	return nil
}

// RevealScriptAddress is a method of the Inscription struct.
// It is responsible for generating the reveal script address.
// It creates a control block, a leaf node, a tap script, and an output key.
// It then generates the taproot address from the output key.
// It returns an error if there is an error in any of the steps.
func (i *Inscription) RevealScriptAddress() error {
	// Create a control block
	controlBlock := &txscript2.ControlBlock{
		InternalKey: i.internalKey,
		LeafVersion: txscript2.BaseLeafVersion,
	}
	i.controlBlock = controlBlock

	// Create a leaf node
	leafNode := txscript.NewBaseTapLeaf(i.revealScript)
	i.tapLeafNode = leafNode

	// Create a tap script
	tapScript := waddrmgr.Tapscript{
		Type: waddrmgr.TapscriptTypeFullTree,
		Leaves: []txscript2.TapLeaf{{
			LeafVersion: txscript2.BaseLeafVersion,
			Script:      i.revealScript,
		}},
		ControlBlock: controlBlock,
	}

	// Generate the output key
	outputKey, err := tapScript.TaprootKey()
	if err != nil {
		return err
	}

	// Determine if the y-coordinate of the output key is odd
	yIsOdd := outputKey.SerializeCompressed()[0] == secp.PubKeyFormatCompressedOdd
	controlBlock.OutputKeyYIsOdd = yIsOdd

	// Generate the taproot address
	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), util.ActiveNet.Params)
	i.revealTxAddress = taprootAddress
	return nil
}

// getUtxo is a method of the Inscription struct.
// It is responsible for getting the unspent transaction outputs (UTXOs) for the wallet.
// It first lists the unspent and locked UTXOs, then filters out the UTXOs that are already used in inscriptions.
// It returns an error if there is an error in any of the steps.
func (i *Inscription) getUtxo() error {
	// List unspent UTXOs
	unspentUtxo, err := i.Wallet().ListUnspent()
	if err != nil {
		return err
	}

	utxo := make([]btcjson.ListUnspentResult, 0)
	for _, v := range unspentUtxo {
		hash, _ := chainhash.NewHashFromStr(v.TxID)
		outpoint := wire.NewOutPoint(hash, v.Vout)
		resp, err := i.options.indexer.Outpoint(outpoint.String())
		if err != nil {
			return err
		}
		if len(resp.Inscriptions) == 0 {
			utxo = append(utxo, v)
		}
	}
	i.utxo = utxo
	return nil
}

// Data is a method of the Inscription struct. It returns the body of the inscription.
func (i *Inscription) Data() []byte {
	return i.body
}

// calculateTxFee is a function that calculates the transaction fee
// for a given transaction and fee rate. It first calculates the weight
// of the transaction, then calculates the fee based on the weight and fee rate.
// If the calculated fee is less than the dust limit, it sets the fee to the dust limit.
// It returns the calculated fee.
func calculateTxFee(tx *wire.MsgTx, feeRate int64) int64 {
	weight := tx.SerializeSizeStripped()*3 + tx.SerializeSize()
	fee := decimal.NewFromInt(int64(weight)).
		Div(decimal.NewFromInt(4)).
		Div(decimal.NewFromInt(1000)). //Ceil().
		Mul(decimal.NewFromInt(feeRate)).IntPart()

	if fee < constants.DustLimit {
		fee = constants.DustLimit
	}
	return fee
}

type Output struct {
	Commit    string `json:"commit"`
	Reveal    string `json:"reveal"`
	TotalFees int64  `json:"total_fees"`
}
