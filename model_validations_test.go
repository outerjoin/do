package do

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInsertableChecks(t *testing.T) {

	// If you insert a field tagged insert:no
	// then an error is generated
	{
		a := struct {
			Field1 string `insert:"no"`
		}{}
		errs2 := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{"field1": "abc"})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field1", errs2[0].Source)

		// And this works even if you pass
		// the address of struct
		{
			errs2 = checkInsertableUpdatableDataInMap(&a, DB_INSERT, map[string]interface{}{"field1": "abc"})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field1", errs2[0].Source)
		}
	}

	// If you insert a field tagged insert:no
	// even at a nested struct level,
	// then an error is generated
	{
		b := struct {
			FieldA struct {
				Field1 string `insert:"no"`
			}
		}{}
		errs2 := checkInsertableUpdatableDataInMap(b, DB_INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field_a.field1", errs2[0].Source)

		// And this works even if you pass
		// the address of struct
		{
			errs2 := checkInsertableUpdatableDataInMap(&b, DB_INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field_a.field1", errs2[0].Source)
		}
	}

	// If you insert a field tagged insert:no
	// even at a nested pointer-struct level,
	// then an error is generated
	{
		c := struct {
			FieldB *struct {
				Field1 string `insert:"no"`
			}
		}{}
		errs2 := checkInsertableUpdatableDataInMap(c, DB_INSERT, map[string]interface{}{"field_b": map[string]interface{}{"field1": "abc"}})
		assert.Equal(t, 1, len(errs2))
		assert.Equal(t, "field_b.field1", errs2[0].Source)

		// And this works even if you pass
		// the address of struct
		{
			errs2 := checkInsertableUpdatableDataInMap(&c, DB_INSERT, map[string]interface{}{"field_b": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 1, len(errs2))
			assert.Equal(t, "field_b.field1", errs2[0].Source)
		}
	}

	// insert:yes field is missing
	// then you get errors
	{
		a := struct {
			Field1 string `insert:"yes"`
		}{}
		errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{})
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, "field1", errs[0].Source)

		// Errors go away wnen data is provided
		{
			errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{"field1": "abc"})
			assert.Equal(t, 0, len(errs))
		}
	}

	// insert:yes field is missing,
	// at a nested struct level
	// then you get errors
	{
		a := struct {
			FieldA struct {
				Field1 string `insert:"yes"`
			}
		}{}
		errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{})
		assert.Equal(t, 1, len(errs))
		assert.Equal(t, "field_a.field1", errs[0].Source)

		// Errors go away wnen data is provided
		{
			errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{"field_a": map[string]interface{}{"field1": "abc"}})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Field is optional to insert
	// Missing it is OK
	{
		a := struct {
			Field1 string `insert:"opt"`
		}{}
		errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{})
		assert.Equal(t, 0, len(errs))
	}

	// Defualt value for insert is [opt]
	{
		a := struct {
			Field1 string /*`insert:"opt"`*/
		}{}
		errs := checkInsertableUpdatableDataInMap(a, DB_INSERT, map[string]interface{}{})
		assert.Equal(t, 0, len(errs))
	}
}

func TestInsertFieldUpdateAction(t *testing.T) {

	// Field is tagged to insert:no
	// There should be no impact on Update action
	{
		a := struct {
			Field1 string `insert:"no"`
		}{}
		errs := checkInsertableUpdatableDataInMap(a, DB_UPDATE, map[string]interface{}{"field1": "abc"})
		assert.Equal(t, 0, len(errs))
	}
}

func TestDefaultsOnInsert(t *testing.T) {

	// For an optional field, if a default value is provided
	// then it should get used
	{
		a := struct {
			Field1 string `default:"abra"`
		}{}
		m := map[string]interface{}{}
		errs := provideDefualts(a, DB_INSERT, m)
		assert.Equal(t, 0, len(errs))
		assert.Equal(t, "abra", m["field1"])
	}

	// When the field is insert:"no",
	// any specified default values should not kick in
	{
		a := struct {
			Field1 string `default:"abra" insert:"no"`
		}{}
		m := map[string]interface{}{}
		errs := provideDefualts(a, DB_INSERT, m)
		assert.Equal(t, 0, len(errs))
		_, found := m["field1"]
		assert.False(t, found)
	}
}

func TestFieldConversion(t *testing.T) {

	var m map[string]interface{}

	{
		a := struct {
			MongoEntity
			Field1 int ``
		}{}
		m = map[string]interface{}{"field1": "12345"}
		errs := convertFieldType(a, DB_INSERT, m)
		_, isInt := m["field1"].(int)
		assert.Equal(t, 0, len(errs))
		assert.True(t, isInt)
	}

	{
		b := struct {
			MongoEntity
			Field1 int64 ``
		}{}
		m = map[string]interface{}{"field1": "12345"}
		errs := convertFieldType(b, DB_INSERT, m)
		_, isInt64 := m["field1"].(int64)
		assert.Equal(t, 0, len(errs))
		assert.True(t, isInt64)
	}
}

func TestTimedFields(t *testing.T) {

	var m map[string]interface{}

	a := struct {
		MongoEntity
		Timed
	}{}
	m = map[string]interface{}{}
	errs := populateTimedFields(a, DB_INSERT, m)
	assert.Equal(t, 0, len(errs))

	_, isTime := m["created_at"].(time.Time)
	assert.True(t, isTime)

	_, isTime = m["updated_at"].(time.Time)
	assert.True(t, isTime)
}

func TestPopulateAutoFields(t *testing.T) {

	{
		a := struct {
			Field1 string `auto:"prefix:555-;uuid"`
		}{}
		m := map[string]interface{}{}
		errs := populateAutoFields(a, DB_INSERT, m)
		assert.Equal(t, 0, len(errs))

		str, found := m["field1"].(string)
		assert.True(t, found)
		assert.True(t, strings.HasPrefix(str, "555-"))
	}

	{
		a := struct {
			Field1 string `auto:"alphanum(5)"`
		}{}
		m := map[string]interface{}{}
		errs := populateAutoFields(a, DB_INSERT, m)
		assert.Equal(t, 0, len(errs))

		str, found := m["field1"].(string)
		assert.True(t, found)
		assert.Regexp(t, `^[a-z0-9A-Z]{5}$`, str)
	}
}

func TestVerifications(t *testing.T) {

	// verify:email
	{
		a := struct {
			Field1 string `verify:"email"`
		}{}
		errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
			"field1": "abc @ def . com",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
				"field1": "abc@def.com",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Regular expression
	// verify:rex(...)
	{
		a := struct {
			Field1 string `verify:"rex(^a+$)"`
		}{}
		errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
			"field1": "abc",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
				"field1": "aaa",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// Complex Regular expression
	{
		a := struct {
			Mobile string `verify:"rex(^[6-9]\\d{9}$)"`
		}{}
		errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
			"mobile": "abc",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
				"mobile": "9977887799",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

	// enum
	{
		a := struct {
			Color string `verify:"enum(green|yellow|red)"`
		}{}
		errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
			"color": "black",
		})
		assert.Equal(t, 1, len(errs))

		// pass
		{
			errs := verifyInputs(a, DB_INSERT, map[string]interface{}{
				"color": "red",
			})
			assert.Equal(t, 0, len(errs))
		}
	}

}

func TestTrim(t *testing.T) {

	a := struct {
		Field1 string `trim:"no"`
		Field2 string ``
	}{}
	m := map[string]interface{}{
		"field1": " 1 ",
		"field2": " 2 ",
	}
	errs := trimFields(a, DB_INSERT, m)
	assert.Equal(t, 0, len(errs))

	assert.Equal(t, " 1 ", m["field1"])
	assert.Equal(t, "2", m["field2"])
}
