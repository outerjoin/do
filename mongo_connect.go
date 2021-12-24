package do

import (
	"context"
	"fmt"

	"github.com/rightjoin/fig"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnect struct {
	ConnStr string // must be present
	DB      string // must be present

	DBSuffix   string // optional
	Coll       string // optional
	CollSuffix string // optional

	// TODO: evaluate reuse of client
	client *mongo.Client
}

func NewMongoConnect() *MongoConnect {
	return &MongoConnect{
		ConnStr: fig.String("database.mongo.connection"),
		DB:      fig.String("database.mongo.db"),
	}
}

func (mc *MongoConnect) CloseClient() {
	if mc.client != nil {
		if err := mc.client.Disconnect(context.TODO()); err != nil {
			log.Error().Msg("unable to close mongodb client connection")
		}
	}
}

func (mc *MongoConnect) Client() *mongo.Client {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mc.ConnStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("connection string", mc.ConnStr).
			Msg("unable to open connection")
		return nil
	}
	return client
}

func (mc *MongoConnect) Database() *mongo.Database {
	if mc.DBSuffix == "" {
		return mc.Client().Database(mc.DB)
	} else {
		return mc.Client().Database(fmt.Sprintf("%s-%s", mc.DB, mc.DBSuffix))
	}
}

func (mc *MongoConnect) Collection(model ...interface{}) *mongo.Collection {

	coll := ""
	if len(model) == 0 {
		coll = mc.Coll
	} else {
		coll = MongoCollectionName(model[0])
	}

	if mc.CollSuffix == "" {
		return mc.Database().Collection(coll)
	} else {
		return mc.Database().Collection(fmt.Sprintf("%s-%s", coll, mc.CollSuffix))
	}
}
