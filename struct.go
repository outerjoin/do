package do

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rightjoin/rutl/conv"
)

type WalkConfig struct {
	Tag string
}

func (wc WalkConfig) FieldKey(fld reflect.StructField) string {
	if wc.Tag != "" {
		key := fld.Tag.Get(wc.Tag)
		if key != "" {
			return key
		}
	}
	return conv.CaseSnake(fld.Name)
}

type FieldOp func(reflect.StructField, Map, ...string) []ErrorPlus

func StructWalk(modelOrType interface{}, c WalkConfig, data Map, action FieldOp, keys ...string) []ErrorPlus {

	errs := []ErrorPlus{}
	stype := TypeOf(modelOrType)
	stype = TypeDereference(stype)

	for i := 0; i < stype.NumField(); i++ {
		fld := stype.Field(i)
		fldName := c.FieldKey(fld)
		fldType := TypeDereference(TypeOf(fld.Type))

		subKeys := keys
		if fldName != "" {
			subKeys = append(keys, fldName)
		}

		if fldType.Kind() == reflect.Struct && !TypeIsTime(fld.Type) {

			if data.HasKey(fldName) {
				dict, isMap := data.GetOr(fldName, false).(map[string]interface{})
				if isMap {
					e := StructWalk(fldType, c, dict, action, subKeys...)
					if e != nil {
						errs = append(errs, e...)
					}
				} else {
					issue := fmt.Sprintf("field '%s' expected dict, but found literal", fldName)
					errs = append(errs, ErrorPlus{Message: issue, Source: strings.Join(subKeys, ".")})
				}
			} else {
				// Pass an empty struct, and use its
				// value if it gets populated
				new := map[string]interface{}{}
				e := StructWalk(fldType, c, new, action, subKeys...)
				if e != nil {
					errs = append(errs, e...)
				}
			}

		} else {
			e := action(fld, data, subKeys...)
			if e != nil {
				errs = append(errs, e...)
			}
		}
	}

	return errs
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// Returns type (never returns ptr type)
func StructGetFieldTypeByJsonKey(modelType interface{}, jsonKey string) (reflect.Type, bool) {

	typesByJson := map[string]reflect.Type{}

	// TODO:
	// cache allFields, so for same struct you don't create it again and again

	mt := TypeOf(modelType)

	for i := 0; i < mt.NumField(); i++ {
		structLoadTypeByKeys(mt.Field(i), "", typesByJson)
	}

	if t, ok := typesByJson[jsonKey]; ok {
		return t, true
	}

	return reflect.TypeOf(nil), false
}

func structLoadTypeByKeys(fld reflect.StructField, prefix string, dest map[string]reflect.Type) {

	t := TypeDereference(fld.Type)
	kind := t.Kind()

	isStruct := kind == reflect.Struct
	isTime := TypeIsTime(t)

	if isTime || kind == reflect.Bool || kind == reflect.String ||
		kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 ||
		kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int || kind == reflect.Int64 ||
		kind == reflect.Float32 || kind == reflect.Float64 {
		currKey := fld.Tag.Get("json")
		if currKey == "" {
			currKey = conv.CaseSnake(fld.Name)
		}
		if prefix != "" {
			currKey = prefix + "." + currKey
		}
		dest[currKey] = t
	} else if isStruct {
		toPass := ""
		currKey := fld.Tag.Get("json")
		if currKey == "" {
			toPass = prefix
		} else {
			if prefix != "" {
				toPass = prefix + "." + currKey
			} else {
				toPass = currKey
			}
		}
		for i := 0; i < t.NumField(); i++ {
			structLoadTypeByKeys(t.Field(i), toPass, dest)
		}
	}
}
