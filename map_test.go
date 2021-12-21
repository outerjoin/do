package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapHasKey(t *testing.T) {
	m := NewMapFromGoMap(map[string]interface{}{"abc": 123})
	assert.Equal(t, true, m.HasKey("abc"))
	assert.Equal(t, false, m.HasKey("def"))
}
