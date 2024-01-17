package index

import (
	"gotest.tools/assert"
	"testing"
)

func TestIndexer_UpdateIndex(t *testing.T) {
	assert.Equal(t, indexer.UpdateIndex(), nil)
}
