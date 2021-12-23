package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructGetFieldTypeByJsonKey(t *testing.T) {
	a := struct {
		FieldC1 string
		FieldC2 int
		Parent  struct {
			FieldP1 string
		} `json:"abc"`
	}{}
	t1, _ := StructGetFieldTypeByJsonKey(a, "abc.field_p1")
	t2, _ := StructGetFieldTypeByJsonKey(a, "field_c1")
	t3, _ := StructGetFieldTypeByJsonKey(a, "field_c2")

	assert.Equal(t, "string", t1.String())
	assert.Equal(t, "string", t2.String())
	assert.Equal(t, "int", t3.String())
}
