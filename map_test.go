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

func TestMapExapndDotKeys(t *testing.T) {

	var out map[string]interface{}

	// no nesting
	out = NewMapFromGoMap(map[string]interface{}{
		"abc": "def",
	}).Unlevel()
	assert.Equal(t, "def", out["abc"])

	// 1 level of nesting
	out = NewMapFromGoMap(map[string]interface{}{
		"abc":     "def",
		"qrs.tuv": "123",
		"qrs.lmn": "456",
	}).Unlevel()
	assert.Equal(t, "def", out["abc"])
	inner1, ok := out["qrs"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "123", inner1["tuv"])
	assert.Equal(t, "456", inner1["lmn"])

	// 2 level of nesting
	out = NewMapFromGoMap(map[string]interface{}{
		"abc":              "def",
		"man.john.doe":     "11",
		"man.john.jacobs":  "22",
		"man.john.fogarty": "33",
		"man.don.bulls":    "44",
	}).Unlevel()
	assert.Equal(t, "def", out["abc"])
	inner2, ok := out["man"].(map[string]interface{})
	assert.True(t, ok)
	inner3a, ok := inner2["john"].(map[string]interface{})
	assert.True(t, ok)
	inner3b, ok := inner2["don"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "11", inner3a["doe"])
	assert.Equal(t, "22", inner3a["jacobs"])
	assert.Equal(t, "33", inner3a["fogarty"])
	assert.Equal(t, "44", inner3b["bulls"])
}
