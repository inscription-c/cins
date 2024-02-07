package constants

import (
	"fmt"
	"regexp"
)

const (
	AppName         = "cins"
	ProtocolId      = "c-ins"
	CInsDescription = "-1"

	DefaultPostage = 10_000
	MaxPostage     = 20_000

	InscriptionIdDelimiter = "i"
	OutpointDelimiter      = ":"
	IdRegexpContent        = `^[a-z0-9]{64}%s\d+$`
)

var (
	InscriptionIdRegexp = regexp.MustCompile(fmt.Sprintf(IdRegexpContent, InscriptionIdDelimiter))
	OutpointRegexp      = regexp.MustCompile(fmt.Sprintf(IdRegexpContent, OutpointDelimiter))
)

type ContentType string

func (t ContentType) Bytes() []byte {
	return []byte(t)
}

func (t ContentType) String() string {
	return string(t)
}

func (t ContentType) MediaType() MediaType {
	for _, media := range Medias {
		if media.ContentType == t {
			return media.MediaType
		}
	}
	return MediaUnknown
}

const (
	ContentTypeCbor             ContentType = "application/cbor"
	ContentTypeJson             ContentType = "application/json"
	ContentTypeOctetStream      ContentType = "application/octet-stream"
	ContentTypePdf              ContentType = "application/pdf"
	ContentTypePgpSignature     ContentType = "application/pgp-signature"
	ContentTypeProtobuf         ContentType = "application/protobuf"
	ContentTypeXJavascript      ContentType = "application/x-javascript"
	ContentTypeYaml             ContentType = "application/yaml"
	ContentTypeAudioFlac        ContentType = "audio/flac"
	ContentTypeAudioMpeg        ContentType = "audio/mpeg"
	ContentTypeAudioWav         ContentType = "audio/wav"
	ContentTypeFrontOtf         ContentType = "font/otf"
	ContentTypeFrontTtf         ContentType = "font/ttf"
	ContentTypeFrontWoff        ContentType = "font/woff"
	ContentTypeFrontWoff2       ContentType = "font/woff2"
	ContentTypeImageApng        ContentType = "image/apng"
	ContentTypeImageAVif        ContentType = "image/avif"
	ContentTypeImageGif         ContentType = "image/gif"
	ContentTypeImageJpeg        ContentType = "image/jpeg"
	ContentTypeImagePng         ContentType = "image/png"
	ContentTypeImageSvgXml      ContentType = "image/svg+xml"
	ContentTypeImageWebp        ContentType = "image/webp"
	ContentTypeModelGltfJson    ContentType = "model/gltf+json"
	ContentTypeModelGltfBinary  ContentType = "model/gltf-binary"
	ContentTypeModelStl         ContentType = "model/stl"
	ContentTypeTextCss          ContentType = "text/css"
	ContentTypeTextHtml         ContentType = "text/html"
	ContentTypeTextHtmlUtf8     ContentType = "text/html;charset=utf-8"
	ContentTypeTextJs           ContentType = "text/javascript"
	ContentTypeTextMarkdown     ContentType = "text/markdown"
	ContentTypeTextMarkdownUtf8 ContentType = "text/markdown;charset=utf-8"
	ContentTypeTextPlain        ContentType = "text/plain"
	ContentTypeTextPlainUtf8    ContentType = "text/plain;charset=utf-8"
	ContentTypeTextXPython      ContentType = "text/x-python"
	ContentTypeVideoMp4         ContentType = "video/mp4"
	ContentTypeVideoWebm        ContentType = "video/webm"
)

type BrotliMode int

const (
	BrotliModeGeneric      BrotliMode = 0
	BrotliModeText         BrotliMode = 1
	BrotliModeFont         BrotliMode = 2
	BrotliForceLsbPrior    BrotliMode = 3
	BrotliForceMsbPrior    BrotliMode = 4
	BrotliForceSignedPrior BrotliMode = 6
)

type MediaType string

func (m MediaType) String() string {
	return string(m)
}

const (
	MediaUnknown    MediaType = "unknown"
	MediaAudio      MediaType = "audio"
	MediaCss        MediaType = "css"
	MediaJavaScript MediaType = "javascript"
	MediaJson       MediaType = "json"
	MediaPython     MediaType = "python"
	MediaYaml       MediaType = "yaml"
	MediaFont       MediaType = "font"
	MediaIframe     MediaType = "iframe"
	MediaImage      MediaType = "image"
	MediaMarkdown   MediaType = "markdown"
	MediaModel      MediaType = "model"
	MediaPdf        MediaType = "pdf"
	MediaText       MediaType = "text"
	MediaVideo      MediaType = "video"
)

type Extension string

const (
	ExtensionCbor  Extension = "cbor"
	ExtensionJson  Extension = "json"
	ExtensionBin   Extension = "bin"
	ExtensionPdf   Extension = "pdf"
	ExtensionAsc   Extension = "asc"
	ExtensionBinPb Extension = "binpb"
	ExtensionYaml  Extension = "yaml"
	ExtensionYml   Extension = "yml"
	ExtensionFlac  Extension = "flac"
	ExtensionMp3   Extension = "mp3"
	ExtensionWav   Extension = "wav"
	ExtensionOtf   Extension = "otf"
	ExtensionTtf   Extension = "ttf"
	ExtensionWoff  Extension = "woff"
	ExtensionWoff2 Extension = "woff2"
	ExtensionApng  Extension = "apng"
	ExtensionGif   Extension = "gif"
	ExtensionJpg   Extension = "jpg"
	ExtensionJpeg  Extension = "jpeg"
	ExtensionPng   Extension = "png"
	ExtensionSvg   Extension = "svg"
	ExtensionWebp  Extension = "webp"
	ExtensionGltf  Extension = "gltf"
	ExtensionGlb   Extension = "glb"
	ExtensionStl   Extension = "stl"
	ExtensionCss   Extension = "css"
	ExtensionHtml  Extension = "html"
	ExtensionJs    Extension = "js"
	ExtensionMd    Extension = "md"
	ExtensionTxt   Extension = "txt"
	ExtensionPy    Extension = "py"
	ExtensionMp4   Extension = "mp4"
	ExtensionWebm  Extension = "webm"
)

type Media struct {
	ContentType ContentType
	BrotliMode  BrotliMode
	MediaType   MediaType
	Extensions  []Extension
}

var Medias = []Media{
	{
		ContentTypeCbor,
		BrotliModeGeneric,
		MediaUnknown,
		[]Extension{ExtensionCbor},
	},
	{
		ContentTypeJson,
		BrotliModeText,
		MediaJson,
		[]Extension{ExtensionJson},
	},
	{
		ContentTypeOctetStream,
		BrotliModeGeneric,
		MediaUnknown,
		[]Extension{ExtensionBin},
	},
	{
		ContentTypePdf,
		BrotliModeGeneric,
		MediaPdf,
		[]Extension{ExtensionPdf},
	},
	{
		ContentTypePgpSignature,
		BrotliModeText,
		MediaText,
		[]Extension{ExtensionAsc},
	},
	{
		ContentTypeProtobuf,
		BrotliModeGeneric,
		MediaUnknown,
		[]Extension{ExtensionBinPb},
	},
	{
		ContentTypeXJavascript,
		BrotliModeText,
		MediaJavaScript,
		[]Extension{},
	},
	{
		ContentTypeYaml,
		BrotliModeText,
		MediaYaml,
		[]Extension{ExtensionYaml, ExtensionYml},
	},
	{
		ContentTypeAudioFlac,
		BrotliModeGeneric,
		MediaAudio,
		[]Extension{ExtensionFlac},
	},
	{
		ContentTypeAudioMpeg,
		BrotliModeGeneric,
		MediaAudio,
		[]Extension{ExtensionMp3},
	},
	{
		ContentTypeAudioWav,
		BrotliModeGeneric,
		MediaAudio,
		[]Extension{ExtensionWav},
	},
	{
		ContentTypeFrontOtf,
		BrotliModeGeneric,
		MediaFont,
		[]Extension{ExtensionOtf},
	},
	{
		ContentTypeFrontTtf,
		BrotliModeGeneric,
		MediaFont,
		[]Extension{ExtensionTtf},
	},
	{
		ContentTypeFrontWoff,
		BrotliModeGeneric,
		MediaFont,
		[]Extension{ExtensionWoff},
	},
	{
		ContentTypeFrontWoff2,
		BrotliModeFont,
		MediaFont,
		[]Extension{ExtensionWoff2},
	},
	{
		ContentTypeImageApng,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{ExtensionApng},
	},
	{
		ContentTypeImageAVif,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{},
	},
	{
		ContentTypeImageGif,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{ExtensionGif},
	},
	{
		ContentTypeImageJpeg,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{ExtensionJpg, ExtensionJpeg},
	},
	{
		ContentTypeImagePng,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{ExtensionPng},
	},
	{
		ContentTypeImageSvgXml,
		BrotliModeText,
		MediaIframe,
		[]Extension{ExtensionSvg},
	},
	{
		ContentTypeImageWebp,
		BrotliModeGeneric,
		MediaImage,
		[]Extension{ExtensionWebp},
	},
	{
		ContentTypeModelGltfJson,
		BrotliModeText,
		MediaModel,
		[]Extension{ExtensionGltf},
	},
	{
		ContentTypeModelGltfBinary,
		BrotliModeGeneric,
		MediaModel,
		[]Extension{ExtensionGlb},
	},
	{
		ContentTypeModelStl,
		BrotliModeGeneric,
		MediaUnknown,
		[]Extension{ExtensionStl},
	},
	{
		ContentTypeTextCss,
		BrotliModeText,
		MediaCss,
		[]Extension{ExtensionCss},
	},
	{
		ContentTypeTextHtml,
		BrotliModeText,
		MediaIframe,
		[]Extension{},
	},
	{
		ContentTypeTextHtmlUtf8,
		BrotliModeText,
		MediaIframe,
		[]Extension{ExtensionHtml},
	},
	{
		ContentTypeTextJs,
		BrotliModeText,
		MediaJavaScript,
		[]Extension{ExtensionJs},
	},
	{
		ContentTypeTextMarkdown,
		BrotliModeText,
		MediaMarkdown,
		[]Extension{},
	},
	{
		ContentTypeTextMarkdownUtf8,
		BrotliModeText,
		MediaMarkdown,
		[]Extension{ExtensionMd},
	},
	{
		ContentTypeTextPlain,
		BrotliModeText,
		MediaText,
		[]Extension{},
	},
	{
		ContentTypeTextPlainUtf8,
		BrotliModeText,
		MediaText,
		[]Extension{ExtensionTxt},
	},
	{
		ContentTypeTextXPython,
		BrotliModeText,
		MediaPython,
		[]Extension{ExtensionPy},
	},
	{
		ContentTypeVideoMp4,
		BrotliModeGeneric,
		MediaVideo,
		[]Extension{ExtensionMp4},
	},
	{
		ContentTypeVideoWebm,
		BrotliModeGeneric,
		MediaVideo,
		[]Extension{ExtensionWebm},
	},
}
