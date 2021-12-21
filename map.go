package do

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Map map[string]interface{}

func NewMap() Map {
	return Map{}
}

func NewMapFromBsonDoc(d bson.D) Map {
	return NewMapFromGoMap(d.Map())
}

func NewMapFromGoMap(dict map[string]interface{}) Map {
	m := Map(dict)
	return m
}

func NewMapFromValues(kv ...interface{}) Map {
	if len(kv) == 1 {
		if kvMap, ok := kv[0].(map[string]interface{}); ok {
			return NewMapFromGoMap(kvMap)
		} else if kvDoc, ok := kv[0].(bson.D); ok {
			return NewMapFromBsonDoc(kvDoc)
		} else {
			//
			panic("only expected map or bson")
		}
	} else {
		// TODO: check if odd number of inputs are given

		// Loop it through and create a map
		tmp := map[string]interface{}{}
		for i := 0; i+1 < len(kv); i = i + 2 {
			tmp[fmt.Sprint(kv[i])] = kv[i+1]
		}
		return NewMapFromGoMap(tmp)
	}
}

func (m Map) HasKey(key string) bool {
	dict := map[string]interface{}(m)
	_, ok := dict[key]
	return ok
}

func (m Map) Get(key string) (interface{}, bool) {
	dict := map[string]interface{}(m)
	val, found := dict[key]
	if found {
		return val, true
	} else {
		return nil, false
	}
}

func (m Map) GetOr(key string, defValue interface{}) interface{} {
	dict := map[string]interface{}(m)
	val, found := dict[key]
	if found {
		return val
	} else {
		return defValue
	}
}
