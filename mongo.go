package do

import (
	"context"
	"reflect"
	"strings"

	"github.com/rightjoin/rutl/conv"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Given a struct, or address of a struct, get the
// appropriate collection name for storing that struct
func MongoCollectionName(model interface{}) string {
	if name, ok := model.(string); ok {
		return name
	}

	// Indirect
	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	// If "CollectionName" method exists, call it
	if _, ok := t.MethodByName("CollectionName"); ok {
		col := v.MethodByName("CollectionName").Call([]reflect.Value{})
		return col[0].String()
	}

	return strings.TrimSpace(conv.CaseSnake(t.Name()))
}

func CloseMongoClient(m *mongo.Client) {
	if err := m.Disconnect(context.TODO()); err != nil {
		log.Error().Msg("unable to close connection")
	}
}
