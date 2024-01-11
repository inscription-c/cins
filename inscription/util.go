package inscription

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/dotbitHQ/insc/constants"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	secp256k12 "github.com/olegabu/go-secp256k1-zkp"
	"path/filepath"
	"strings"
)

func GetXOnlyPublicKey() (*secp256k12.XonlyPubkey, error) {
	sk := secp256k12.Random256()
	kp, err := secp256k12.KeypairCreate(secp256k12.SharedContext(secp256k12.ContextBoth), sk[:])
	if err != nil {
		return nil, err
	}
	noneCtx, err := secp256k12.ContextCreate(secp256k12.ContextNone)
	if err != nil {
		return nil, err
	}
	xopk, _, err := secp256k12.KeypairXonlyPubkey(noneCtx, kp)
	if err != nil {
		return nil, err
	}

	return xopk, nil
}

func GetTapScript(internalKey *btcec.PublicKey, script []byte) (*waddrmgr.Tapscript, error) {
	tapScript := &waddrmgr.Tapscript{
		Type: waddrmgr.TapscriptTypeFullTree,
		Leaves: []txscript.TapLeaf{
			{
				LeafVersion: txscript.BaseLeafVersion,
				Script:      script,
			},
		},
		ControlBlock: &txscript.ControlBlock{
			InternalKey: internalKey,
		},
	}
	return tapScript, nil
}

func ContentTypeForPath(path string) (*constants.Media, error) {
	ext := constants.Extension(strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")))
	if ext == constants.ExtensionMp4 {
		ok, err := CheckMp4Codec(path)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("mp4 file codec must be h264")
		}
	}
	for idx := range constants.Medias {
		media := constants.Medias[idx]
		for _, v := range media.Extensions {
			if v == ext {
				return &media, nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported file extension for `%s`", ext)
}

func CheckMp4Codec(path string) (bool, error) {
	file, err := avutil.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	streams, err := file.Streams()
	if err != nil {
		return false, err
	}

	for _, stream := range streams {
		if _, ok := stream.(av.VideoCodecData); !ok {
			return false, nil
		}
		if stream.Type() != av.H264 {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}
