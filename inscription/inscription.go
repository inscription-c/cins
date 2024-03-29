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
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	txscript2 "github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/go-playground/validator/v10"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/inscription/log"
	"github.com/inscription-c/cins/pkg/indexer"
	"github.com/inscription-c/cins/pkg/util"
	"github.com/inscription-c/cins/pkg/util/txscript"
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
	commitTx, revealTx *wire.MsgTx

	// priKey is the private key for the transaction.
	priKey *btcec.PrivateKey

	// internalKey is the internal key for the transaction.
	internalKey *btcec.PublicKey

	// revealScript is the reveal script for the transaction.
	revealScript []byte

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

	// indexer is the indexer for the inscription.
	indexer indexer.IndexerInterface
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

func WithIndexer(indexer indexer.IndexerInterface) func(*options) {
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
	media, err := util.ContentTypeForPath(path)
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewFromData(body, media, inputOpts...)
}

func NewFromData(body []byte, media *constants.Media, inputOpts ...Option) (*Inscription, error) {
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

	// Set the content type of the Inscription
	inscription.Header.ContentType = media.ContentType

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
	backendVersion, err := i.Wallet().BackendVersion()
	if err != nil {
		return err
	}

	var feeRate float64
	if backendVersion == rpcclient.Btcd {
		feeRate, err = i.Wallet().EstimateFee(10)
		if err != nil {
			return err
		}
	} else {
		var resp *btcjson.EstimateSmartFeeResult
		resp, err = i.Wallet().EstimateSmartFee(10, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		if len(resp.Errors) > 0 {
			return errors.New(gconv.String(resp.Errors))
		}
		feeRate = *resp.FeeRate
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
		hash, err := chainhash.NewHashFromStr(v.TxID)
		if err != nil {
			return err
		}
		txIn := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: v.Vout,
		}, nil, nil)
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
	change := inTotal - outTotal - CalculateTxFee(commitTx, i.feeRate)
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
	fee := CalculateTxFee(commitTx, i.feeRate)
	i.totalFee += fee
	change = inTotal - outTotal - fee
	commitTx.TxOut[len(commitTx.TxOut)-1].Value = change
	if change < constants.DustLimit {
		commitTx.TxOut = commitTx.TxOut[:len(commitTx.TxOut)-1]
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
	i.internalKey = i.priKey.PubKey()

	// Start building the reveal script
	// Append the inscription content to the script builder
	revealScript, err := InscriptionToScript(
		i.internalKey,
		i.Header,
		i.Body,
	)
	if err != nil {
		return err
	}
	i.revealScript = revealScript

	// Generate the script address
	controlBlock, taprootAddress, err := RevealScriptAddress(i.internalKey, revealScript)
	if err != nil {
		return err
	}
	i.controlBlock = controlBlock
	i.revealTxAddress = taprootAddress
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
	i.revealFee = CalculateTxFee(revealTx, i.feeRate)
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
	priKeyMap := make(map[string]*btcutil.WIF)
	prevPkScripts := make([][]byte, 0)
	prevPkScriptsMap := make(map[string][]byte)
	inputValues := make([]btcutil.Amount, 0)

	for j := 0; j < len(i.utxo); j++ {
		utxo := i.utxo[j]
		address, err := btcutil.DecodeAddress(utxo.Address, util.ActiveNet.Params)
		if err != nil {
			return err
		}
		wif, err := i.Wallet().DumpPrivKey(address)
		if err != nil {
			return err
		}
		priKeyMap[address.String()] = wif
		pkScript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return err
		}
		prevPkScriptsMap[address.String()] = pkScript
		prevPkScripts = append(prevPkScripts, pkScript)

		amount, err := btcutil.NewAmount(utxo.Amount)
		if err != nil {
			return err
		}
		inputValues = append(inputValues, amount)
	}

	if err := txauthor.AddAllInputScripts(i.commitTx, prevPkScripts, inputValues, secretSource{
		priKeys: priKeyMap,
		scripts: prevPkScriptsMap,
	}); err != nil {
		return err
	}
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
	signHash, err := txscript.CalcTapScriptSignatureHash(sigHashes, txscript.SigHashDefault, i.revealTx, 0, prevFetcher, txscript.NewBaseTapLeaf(i.revealScript))
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

// InscriptionToScript is a method of the Inscription struct. It is
// responsible for appending the reveal script to the script builder. It adds the
// protocol ID, content type, metadata, content encoding, and body to the script builder.
// It returns an error if there is an error in any of the steps.
func InscriptionToScript(
	internalKey *btcec.PublicKey,
	header Header,
	body util.Protocol,
) ([]byte, error) {
	// This block of code is part of the appendRevealScriptToBuilder method of the Inscription struct.
	// It is responsible for appending the reveal script to the script builder.
	// Append check sign op to the script
	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddData(schnorr.SerializePubKey(internalKey))
	scriptBuilder.AddOp(txscript.OP_CHECKSIG)

	// Start building the reveal script
	// Add the initial operations and the protocol ID to the script builder
	scriptBuilder.
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte(constants.ProtocolId)).
		AddData([]byte(constants.CInsDescription)).
		AddData(header.CInsDescription.Data()).
		AddOp(txscript.OP_1).
		AddData(header.ContentType.Bytes())

	// If metadata exists, add it to the script builder
	// The metadata is divided into chunks of 520 bytes and each chunk is added to the script builder
	if header.Metadata != nil && header.Metadata.Len() > 0 {
		for {
			data, err := header.Metadata.Chunks(520)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if err == io.EOF {
				break
			}
			scriptBuilder.AddOp(txscript.OP_5)
			scriptBuilder.AddData(data)
		}
	}

	// If content encoding exists, add it to the script builder
	if header.ContentEncoding != "" {
		scriptBuilder.AddOp(txscript.OP_9)
		scriptBuilder.AddData([]byte(header.ContentEncoding))
	}

	// If body exists, add it to the script builder
	// The body is divided into chunks of 520 bytes and each chunk is added to the script builder
	if body.Len() > 0 {
		scriptBuilder.AddOp(txscript.OP_0)
		for {
			d, err := body.Chunks(520)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if err == io.EOF {
				break
			}
			scriptBuilder.AddData(d)
		}
	}

	// Add the end if operation to the script builder
	scriptBuilder.AddOp(txscript.OP_ENDIF)
	revealScript, err := scriptBuilder.Script()
	if err != nil {
		return nil, err
	}
	return revealScript, nil
}

// RevealScriptAddress is a method of the Inscription struct.
// It is responsible for generating the reveal script address.
// It creates a control block, a leaf node, a tap script, and an output key.
// It then generates the taproot address from the output key.
// It returns an error if there is an error in any of the steps.
func RevealScriptAddress(internalKey *btcec.PublicKey, revealScript []byte) (
	controlBlock *txscript2.ControlBlock,
	taprootAddress *btcutil.AddressTaproot,
	err error) {
	// Create a control block
	controlBlock = &txscript2.ControlBlock{
		InternalKey: internalKey,
		LeafVersion: txscript2.BaseLeafVersion,
	}

	// Create a tap script
	tapScript := waddrmgr.Tapscript{
		Type: waddrmgr.TapscriptTypeFullTree,
		Leaves: []txscript2.TapLeaf{{
			LeafVersion: txscript2.BaseLeafVersion,
			Script:      revealScript,
		}},
		ControlBlock: controlBlock,
	}

	// Generate the output key
	var outputKey *btcec.PublicKey
	outputKey, err = tapScript.TaprootKey()
	if err != nil {
		return
	}

	// Determine if the y-coordinate of the output key is odd
	yIsOdd := outputKey.SerializeCompressed()[0] == secp.PubKeyFormatCompressedOdd
	controlBlock.OutputKeyYIsOdd = yIsOdd

	// Generate the taproot address
	taprootAddress, err = btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), util.ActiveNet.Params)
	return
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

// backupPrivKey is a method of the Inscription struct.
// It is responsible for backing up the private key of the Inscription.
// If the noBackup flag is set, it returns immediately.
// Otherwise, it unlocks the wallet, generates a Wallet Import Format (WIF) from the private key,
// imports the WIF into the wallet, and then locks the wallet.
// It returns an error if there is an error in any of the steps.
func (i *Inscription) backupPrivKey() error {
	if noBackup {
		return nil
	}
	wif, err := btcutil.NewWIF(i.priKey, util.ActiveNet.Params, true)
	if err != nil {
		return err
	}
	if err := i.Wallet().ImportPrivKey(wif); err != nil {
		return err
	}
	return nil
}

// Data is a method of the Inscription struct. It returns the body of the inscription.
func (i *Inscription) Data() []byte {
	return i.body
}

// CalculateTxFee is a function that calculates the transaction fee
// for a given transaction and fee rate. It first calculates the weight
// of the transaction, then calculates the fee based on the weight and fee rate.
// If the calculated fee is less than the dust limit, it sets the fee to the dust limit.
// It returns the calculated fee.
func CalculateTxFee(tx *wire.MsgTx, feeRate int64) int64 {
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

type secretSource struct {
	priKeys map[string]*btcutil.WIF
	scripts map[string][]byte
}

func (s secretSource) GetKey(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
	priKey, ok := s.priKeys[addr.String()]
	if !ok {
		return nil, false, nil
	}
	return priKey.PrivKey, priKey.CompressPubKey, nil
}

func (s secretSource) GetScript(addr btcutil.Address) ([]byte, error) {
	script, ok := s.scripts[addr.String()]
	if !ok {
		return nil, errors.New("no script")
	}
	return script, nil
}

func (s secretSource) ChainParams() *chaincfg.Params {
	return util.ActiveNet.Params
}

type Output struct {
	Commit    string `json:"commit"`
	Reveal    string `json:"reveal"`
	TotalFees int64  `json:"total_fees"`
}
