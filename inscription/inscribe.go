package inscription

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/waddrmgr"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/dotbitHQ/insc/client"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/constants"
	"github.com/dotbitHQ/insc/index"
	"github.com/dotbitHQ/insc/internal/signal"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var (
	InsufficientBalanceError = errors.New("InsufficientBalanceError")

	Cmd = &cobra.Command{
		Use:   "inscribe",
		Short: "inscription casting",
		Run: func(cmd *cobra.Command, args []string) {
			if err := inscribe(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			signal.SimulateInterrupt()
			<-signal.InterruptHandlersDone
		},
	}
)

func inscribe() error {
	if err := configCheck(); err != nil {
		return err
	}

	db, err := index.Open(config.IndexDir)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		if err := db.Close(); err != nil {
			log.Error("db.Close", "err", err)
		}
	})

	inscription, err := FromPath(config.FilePath)
	if err != nil {
		return err
	}
	log.Info(string(inscription.Data()))

	cli := client.RPC()
	unspentUtxo, err := cli.ListUnspent(&client.ListUnspentReq{})
	if err != nil {
		return err
	}
	lockedUtxos, err := cli.ListLockUnspent()
	if err != nil {
		return err
	}

	utoxTotal := make([]client.OutPoint, 0)
	for _, v := range unspentUtxo {
		utoxTotal = append(utoxTotal, v.OutPoint)
	}
	for _, v := range lockedUtxos {
		utoxTotal = append(utoxTotal, v.OutPoint)
	}
	log.Info(gconv.String(utoxTotal))

	// get wallet inscriptions utxo
	walletInscriptions, err := index.GetInscriptionByOutPoints(utoxTotal)
	if err != nil {
		return err
	}
	log.Info(walletInscriptions)

	availableUtxo := make([]client.ListUnspentResp, 0)
	for _, v := range unspentUtxo {
		if _, ok := walletInscriptions[v.OutPoint.String()]; ok {
			continue
		}
		availableUtxo = append(availableUtxo, v)
	}

	commitTx, revealTx, err := createInscriptionTx(inscription, availableUtxo)
	if err != nil {
		return err
	}
	rawSignedCommitTx, err := signCommitTx(commitTx, availableUtxo)
	if err != nil {
		return err
	}
	log.Info("rawSignedCommitTx", rawSignedCommitTx)

	revealTxBuf := bytes.NewBufferString("")
	if err := revealTx.Serialize(revealTxBuf); err != nil {
		return err
	}
	rawSignedRevealTx := hex.EncodeToString(revealTxBuf.Bytes())
	log.Info("rawSignedRevealTx", rawSignedRevealTx)

	commitTxHash, err := client.RPC().SendRawTransaction(rawSignedCommitTx)
	if err != nil {
		return err
	}
	log.Info("commitTxSendSuccess", commitTxHash)

	revealTxHash, err := client.RPC().SendRawTransaction(rawSignedRevealTx)
	if err != nil {
		return err
	}
	log.Info("revealTxSendSuccess", revealTxHash)

	return nil
}

func createInscriptionTx(
	inscription *Inscription,
	unspentUtxo []client.ListUnspentResp) (*wire.MsgTx, *wire.MsgTx, error) {

	priKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	internalKey := priKey.PubKey()
	if err != nil {
		return nil, nil, err
	}

	feeRate, err := client.RPC().EstimateFee(10)
	if err != nil {
		return nil, nil, err
	}
	revealTx, tapAddress, tapLeafNode, revealTxFee, err := buildRevealTx(inscription, internalKey, feeRate)
	if err != nil {
		return nil, nil, err
	}
	commitTx, vout, err := buildCommitTx(unspentUtxo, tapAddress, feeRate, revealTxFee)
	if err != nil {
		return nil, nil, err
	}
	if err := signRevealTx(commitTx, revealTx, priKey, tapLeafNode, vout); err != nil {
		return nil, nil, err
	}
	return commitTx, revealTx, nil
}

func buildCommitTx(
	availableUtxo []client.ListUnspentResp,
	recipientAddress *btcutil.AddressTaproot,
	feeRate uint64,
	revealFee int64) (*wire.MsgTx, int, error) {

	var inTotal, outTotal int64
	commitTx := wire.NewMsgTx(2)
	for _, v := range availableUtxo {
		outPoint, err := wire.NewOutPointFromString(v.OutPoint.Outpoint())
		if err != nil {
			return nil, 0, err
		}
		txIn := wire.NewTxIn(outPoint, nil, nil)
		commitTx.AddTxIn(txIn)
		inTotal += int64(v.Amount.Sat())
	}

	vout := 0
	outTotal += int64(config.Postage) + revealFee

	netParams := &chaincfg.MainNetParams
	if config.Testnet {
		netParams = &chaincfg.TestNet3Params
	}
	decodeRecipientAddress, err := btcutil.DecodeAddress(recipientAddress.String(), netParams)
	if err != nil {
		return nil, 0, err
	}
	recipientScript, err := txscript.PayToAddrScript(decodeRecipientAddress)
	if err != nil {
		return nil, 0, err
	}
	commitTx.AddTxOut(wire.NewTxOut(int64(config.Postage)+revealFee, recipientScript))

	// change
	charge := inTotal - outTotal - calculateTxFee(len(commitTx.TxIn), len(commitTx.TxOut), feeRate)
	if charge < 0 {
		return nil, 0, InsufficientBalanceError
	}

	charge = inTotal - outTotal - calculateTxFee(len(commitTx.TxIn), len(commitTx.TxOut)+1, feeRate)
	if charge >= constants.DustLimit {
		commitTxChangeAddr, err := client.RPC().GetChangeAddress(constants.AddressTypeP2shSegWit)
		if err != nil {
			return nil, 0, err
		}
		decodeAddress, err := btcutil.DecodeAddress(commitTxChangeAddr, netParams)
		if err != nil {
			return nil, 0, err
		}
		script, err := txscript.PayToAddrScript(decodeAddress)
		if err != nil {
			return nil, 0, err
		}
		commitTx.AddTxOut(wire.NewTxOut(charge, script))
	}
	return commitTx, vout, nil
}

func buildRevealTx(
	inscription *Inscription,
	internalKey *btcec.PublicKey,
	feeRate uint64) (revealTx *wire.MsgTx, tapAddress *btcutil.AddressTaproot, leafNode txscript.TapLeaf, fee int64, err error) {

	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(internalKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	if err = appendRevealScriptToBuilder(builder, inscription); err != nil {
		return
	}
	var revealScript []byte
	revealScript, err = builder.Script()
	if err != nil {
		return
	}
	controlBlock := &txscript.ControlBlock{
		OutputKeyYIsOdd: true,
		InternalKey:     internalKey,
		LeafVersion:     txscript.BaseLeafVersion,
	}
	leafNode = txscript.NewBaseTapLeaf(revealScript)
	tapScript := waddrmgr.Tapscript{
		Type:         waddrmgr.TapscriptTypeFullTree,
		Leaves:       []txscript.TapLeaf{leafNode},
		ControlBlock: controlBlock,
	}

	var outputKey *btcec.PublicKey
	outputKey, err = tapScript.TaprootKey()
	if err != nil {
		return
	}
	controlBlock.OutputKeyYIsOdd = outputKey.SerializeCompressed()[0] == secp.PubKeyFormatCompressedOdd

	netParams := &chaincfg.MainNetParams
	if config.Testnet {
		netParams = &chaincfg.TestNet3Params
	}
	tapAddress, err = btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(outputKey),
		netParams,
	)
	if err != nil {
		return
	}
	log.Info("tapAddress", tapAddress.String())

	revealTxWitness := make([][]byte, 0)
	revealTxWitness = append(revealTxWitness, make([]byte, 64))
	revealTxWitness = append(revealTxWitness, revealScript)

	var controlBlockBytes []byte
	controlBlockBytes, err = controlBlock.ToBytes()
	if err != nil {
		return
	}
	revealTxWitness = append(revealTxWitness, controlBlockBytes)
	revealTxIn := &wire.TxIn{
		Witness:  revealTxWitness,
		Sequence: 0xFFFFFFFD,
	}

	// output
	var desAddress btcutil.Address
	desAddress, err = btcutil.DecodeAddress(strings.TrimSpace(config.Destination), netParams)
	if err != nil {
		return
	}
	var desScript []byte
	desScript, err = txscript.PayToAddrScript(desAddress)
	if err != nil {
		return
	}
	revealTxOutput := wire.NewTxOut(int64(config.Postage), desScript)

	// reveal tx
	revealTx = wire.NewMsgTx(2)
	revealTx.AddTxIn(revealTxIn)
	revealTx.AddTxOut(revealTxOutput)
	fee = calculateTxFee(len(revealTx.TxIn), len(revealTx.TxOut), feeRate)
	return
}

func signCommitTx(commitTx *wire.MsgTx, availableUtxo []client.ListUnspentResp) (string, error) {
	if err := client.RPC().WalletPassphrase(config.WalletPass, 60); err != nil {
		return "", err
	}

	priKeyMap := make(map[string]string)
	feature := txscript.NewMultiPrevOutFetcher(nil)
	for i := 0; i < len(availableUtxo); i++ {
		utxo := availableUtxo[i]
		priKey, err := client.RPC().DumpPriKey(utxo.Address)
		if err != nil {
			return "", err
		}
		priKeyMap[utxo.Outpoint()] = priKey

		outpoint, err := wire.NewOutPointFromString(utxo.Outpoint())
		if err != nil {
			return "", err
		}
		pkscript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return "", err
		}
		feature.AddPrevOut(*outpoint, wire.NewTxOut(int64(utxo.Amount.Sat()), pkscript))
	}
	sigHashes := txscript.NewTxSigHashes(commitTx, feature)

	netParams := &chaincfg.MainNetParams
	if config.Testnet {
		netParams = &chaincfg.TestNet3Params
	}

	for i := 0; i < len(availableUtxo); i++ {
		utxo := availableUtxo[i]
		pkScript, prikKey, _, compress, err := hexPrivateKeyToScript(utxo.Address, netParams, priKeyMap[utxo.Outpoint()])
		if err != nil {
			return "", err
		}
		witness, err := txscript.WitnessSignature(commitTx, sigHashes, i, int64(utxo.Amount.Sat()), pkScript, txscript.SigHashAll, prikKey, compress)
		if err != nil {
			return "", err
		}
		commitTx.TxIn[i].Witness = witness
		sign := hex.EncodeToString(witness[0])
		pk := hex.EncodeToString(witness[1])
		log.Info("sign", len(witness[0]), sign)
		log.Info("pk", len(witness[1]), pk)
	}

	buf := bytes.NewBufferString("")
	if err := commitTx.Serialize(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func signRevealTx(commitTx, revealTx *wire.MsgTx, priKey *btcec.PrivateKey, tapLeafNode txscript.TapLeaf, vout int) error {
	commitHash := commitTx.TxHash()
	revealTx.TxIn[vout].PreviousOutPoint = *wire.NewOutPoint(&commitHash, uint32(vout))

	prevFetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
		revealTx.TxIn[vout].PreviousOutPoint: {
			Value:    commitTx.TxOut[vout].Value,
			PkScript: commitTx.TxOut[vout].PkScript,
		},
	})
	sigHashes := txscript.NewTxSigHashes(revealTx, prevFetcher)
	signHash, err := txscript.CalcTapscriptSignaturehash(sigHashes, txscript.SigHashDefault, revealTx, 0, prevFetcher, tapLeafNode)
	if err != nil {
		return err
	}
	signature, err := schnorr.Sign(priKey, signHash)
	if err != nil {
		return err
	}
	sig := signature.Serialize()
	revealTx.TxIn[vout].Witness[0] = sig
	return nil
}

func appendRevealScriptToBuilder(builder *txscript.ScriptBuilder, inscription *Inscription) error {
	builder.
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte(constants.ProtocolId)).
		AddOp(txscript.OP_1).
		AddData(inscription.Header.ContentType.Bytes())

	if inscription.Header.Metadata.Len() > 0 {
		for {
			data, err := inscription.Header.Metadata.Chunks(520)
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			builder.AddOp(txscript.OP_5)
			builder.AddData(data)
		}
	}

	if inscription.Header.ContentEncoding != "" {
		builder.AddOp(txscript.OP_9)
		builder.AddData([]byte(inscription.Header.ContentEncoding))
	}

	if inscription.Body.Len() > 0 {
		for {
			body, err := inscription.Body.Chunks(520)
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			builder.AddData([]byte{constants.BodyTag})
			builder.AddData(body)
		}
	}
	builder.AddOp(txscript.OP_ENDIF)
	return nil
}

func calculateTxFee(in, out int, feeRate uint64) int64 {
	fee := decimal.NewFromFloat(10.5).
		Add(decimal.NewFromInt(int64(68 * in))).
		Add(decimal.NewFromInt(int64(31 * out))).
		Mul(decimal.NewFromInt(int64(feeRate))).
		Div(decimal.NewFromInt(1000)).IntPart()
	if fee < constants.DustLimit {
		fee = constants.DustLimit
	}
	return fee
}

func hexPrivateKeyToScript(addr string, params *chaincfg.Params, privateKeyStr string) (pkScript []byte, privateKey *btcec.PrivateKey, pk *btcec.PublicKey, compress bool, e error) {
	scriptAddr, err := btcutil.DecodeAddress(addr, params)
	if err != nil {
		e = fmt.Errorf("btcutil.DecodeAddress err: %s", err.Error())
		return
	}
	pkScript, err = txscript.PayToAddrScript(scriptAddr)
	if err != nil {
		e = fmt.Errorf("txscript.PayToAddrScript err: %s", err.Error())
		return
	}

	var wif *btcutil.WIF
	wif, err = btcutil.DecodeWIF(privateKeyStr)
	if err != nil {
		return
	}
	compress = wif.CompressPubKey
	privateKey = wif.PrivKey
	if wif.CompressPubKey {
		pk = privateKey.PubKey()
	} else {
		pk = privateKey.PubKey()
	}
	return
}
