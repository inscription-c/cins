package util

import (
	"errors"
	"fmt"
	"github.com/inscription-c/cins/constants"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"path/filepath"
	"strings"
)

// ContentTypeForPath is a function that takes a file path as input and returns the media type of the file and an error.
// It first checks the file extension and if it's mp4, it checks the codec of the file.
// If the codec is not h264, it returns an error.
// If the file extension is found in the list of supported media types, it returns the media type.
// If the file extension is not supported, it returns an error.
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

// CheckMp4Codec is a function that takes a file path as input and returns a boolean indicating whether the file's codec is h264 and an error.
// It first opens the file and gets the streams.
// It then checks each stream to see if it's a video stream and if the codec is h264.
// If the stream is not a video stream or the codec is not h264, it returns false.
// If all streams are video streams and the codec is h264, it returns true.
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
