package inscription

import (
	"bytes"
	"github.com/ugorji/go/codec"
	"gotest.tools/assert"
	"testing"
)

func TestMetadata(t *testing.T) {
	handle := &codec.CborHandle{}
	m := map[string]interface{}{
		"k1": "v1",
		"k2": 1,
		"k3": []string{"v1", "v2"},
		"k4": map[string]interface{}{
			"kk1": "vv1",
		},
	}

	encBuf := bytes.NewBufferString("")
	assert.Equal(t, codec.NewEncoder(encBuf, handle).Encode(m), nil)
	t.Log(encBuf.String())

	var metadata interface{}
	assert.Equal(t, codec.NewDecoder(encBuf, handle).Decode(&metadata), nil)
	t.Log(metadata)
}
