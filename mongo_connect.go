package do

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/k0kubun/pp"
	"github.com/rightjoin/fig"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
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

	if mc.client != nil {
		return mc.client
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mc.ConnStr))
	if err != nil {
		log.Error().
			Err(err).
			Str("connection string", mc.ConnStr).
			Msg("unable to open connection")
		return nil
	}
	mc.client = client
	return mc.client
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

func (mc *MongoConnect) Query(model interface{}, addrSlice interface{}, opts ...QueryOptions) (int, error) {

	var opt = QueryOptions{
		Query: bson.D{},
		Sort:  bson.M{},
	}
	if len(opts) != 0 {
		opt = opts[0]
	}

	if !opt.Paginate {
		cursor, err := mc.Collection(model).Find(context.Background(), opt.Query, &options.FindOptions{
			Skip:  P_int64(int64(opt.Skip)),
			Limit: P_int64(int64(opt.Limit)),
			Sort:  opt.Sort,
		})
		if err != nil {
			return 0, err
		}

		err = cursor.All(context.Background(), addrSlice)
		if err != nil {
			return 0, err
		}

		// TODO: how to count
		return 0, nil
	}

	// Find total number of records in DB
	total, err := mc.Collection(model).CountDocuments(context.Background(), opt.Query)
	if err != nil {
		pp.Println("01")
		return 0, err
	}

	if opt.Page <= 0 {
		opt.Page = 1
	}
	max := fig.IntOr(25, "pagination.chunk")
	if opt.Chunk < 1 || opt.Chunk > max {
		opt.Chunk = max
	}

	// Find records
	cursor, err := mc.Collection(model).Find(context.Background(), opt.Query, &options.FindOptions{
		Skip:  P_int64(int64((opt.Page - 1) * opt.Chunk)),
		Limit: P_int64(int64(opt.Chunk)),
		Sort:  opt.Sort,
	})
	if err != nil {
		pp.Println("02")
		return 0, err
	}

	err = cursor.All(context.Background(), addrSlice)
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

type QueryOptions struct {
	Query interface{}
	Sort  interface{}
	Skip  int
	Limit int

	// When Paginate EQ true, then 'skip' and 'limit' are
	// essentially ignored. Otherwise 'page' and 'chunk'
	// get ignored
	Paginate bool
	Page     int
	Chunk    int
}

func (mc *MongoConnect) Transactionally(doAction func(sessCtx mongo.SessionContext) error) error {

	// Session
	session, err := mc.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	err = mongo.WithSession(context.Background(), session, func(sessCtx mongo.SessionContext) error {

		var output error

		// Start the transaction
		if output = session.StartTransaction(); output != nil {
			return output
		}

		// Do the work
		output = doAction(sessCtx)
		if output != nil {
			return output
		}

		// Commit the transaction
		if output = session.CommitTransaction(sessCtx); output != nil {
			return output
		}

		return nil
	})
	if err != nil {
		abortErr := session.AbortTransaction(context.Background())
		if abortErr != nil {
			// TODO: log abort error
		}
		return err
	}

	return nil
}

func (mc *MongoConnect) InsertForm(addrObject interface{}, inputs Map, sessCtx ...mongo.SessionContext) []ErrorPlus {

	// Validate inputs for validation errors before sending
	// inputs to DB
	errors := ModelValidateInputs(addrObject, DB_INSERT, inputs)
	if len(errors) > 0 {
		return errors
	}

	var manyErrs []ErrorPlus
	var err error
	fn := func(sessCtx mongo.SessionContext) error {

		// Insert
		result, err := mc.Collection(addrObject).InsertOne(sessCtx, inputs)
		if err != nil {
			return err
		}

		// Read
		err = mc.Collection(addrObject).FindOne(sessCtx, bson.M{"_id": result.InsertedID}).Decode(addrObject)
		if err != nil {
			return err
		}

		// Validate post inserting to DB
		manyErrs = ModelValidateObject(addrObject)
		if len(manyErrs) > 0 {
			return manyErrs[0]
		}

		return nil
	}

	// If there is already a context available
	// then run it under it.
	// Otherwise, start a new transaction.
	if len(sessCtx) > 0 {
		err = fn(sessCtx[0])
	} else {
		err = mc.Transactionally(fn)
	}

	if err != nil {
		if manyErrs != nil && len(manyErrs) > 0 && manyErrs[0] == err {
			return manyErrs
		} else {
			return []ErrorPlus{{Message: err.Error()}}
		}
	}

	return []ErrorPlus{}
}

func (mc *MongoConnect) InsertObject(addrObject interface{}, object interface{}, sessCtx ...mongo.SessionContext) []ErrorPlus {

	// Encode object (data passed) into map
	b, err := json.Marshal(object)
	if err != nil {
		return []ErrorPlus{{Message: err.Error()}}
	}

	// Populate a map from json bytes
	data := map[string]interface{}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return []ErrorPlus{{Message: err.Error()}}
	}

	// Remove any "auto" keys from this map
	ot := TypeOf(addrObject)
	ot = TypeDereference(ot)
	wc := WalkConfig{Tag: "json"}
	for i := 0; i < ot.NumField(); i++ {
		ft := ot.Field(i)
		if ft.Tag.Get("auto") != "" {
			delete(data, wc.FieldKey(ft))
		}
	}

	// Pass on this map to InsertForm to write to DB
	// after performing necessary Validations()
	return mc.InsertForm(addrObject, data, sessCtx...)
}

func (mc *MongoConnect) UpdateForm(addrObject interface{}, queryOne interface{}, inputs Map, sessCtx ...mongo.SessionContext) []ErrorPlus {

	// Validate inputs for validation errors before sending
	// inputs to DB
	errors := ModelValidateInputs(addrObject, DB_UPDATE, inputs)
	if len(errors) > 0 {
		return errors
	}

	// Are any of StateMachine fields being changed?
	coll := MongoCollectionName(addrObject)
	smFields := StateMachine{}.GetStateMachineFieldNames(addrObject)
	smFieldChanged := false
	for _, f := range smFields {
		if inputs.HasKey(f) {
			smFieldChanged = true
			break
		}
	}

	var manyErrs []ErrorPlus
	var err error
	smPreValues := map[string]string{}

	fn := func(sessCtx mongo.SessionContext) error {

		// If state machine field is changed, then we need to
		// fetcht the previous state of object as well
		if smFieldChanged {
			err := mc.Collection(addrObject).FindOne(sessCtx, queryOne).Decode(addrObject)
			if err != nil {
				return err
			}

			// Object to []byte
			b, err := json.Marshal(reflect.ValueOf(addrObject).Elem().Interface())
			if err != nil {
				return err
			}

			// []byte to map
			var objMap map[string]interface{}
			err = json.Unmarshal(b, &objMap)
			if err != nil {
				return err
			}

			// Loop map and save values of state machine fields before
			// update is made
			for _, f := range smFields {
				if pre, ok := objMap[f]; ok {
					smPreValues[f] = pre.(string)
				}
			}
		}

		// Update
		_, err := mc.Collection(addrObject).UpdateOne(sessCtx, queryOne, map[string]interface{}{"$set": inputs})
		if err != nil {
			return err
		}
		// TODO: matched count and modified count check??

		// Read again (after update)
		err = mc.Collection(addrObject).FindOne(sessCtx, queryOne).Decode(addrObject)
		if err != nil {
			return err
		}

		// Validate post inserting to DB
		manyErrs = ModelValidateObject(addrObject)
		if len(manyErrs) > 0 {
			return manyErrs[0]
		}

		if smFieldChanged {
			// Object to []byte
			b, err := json.Marshal(reflect.ValueOf(addrObject).Elem().Interface())
			if err != nil {
				return err
			}

			// []byte to map
			var objMap map[string]interface{}
			err = json.Unmarshal(b, &objMap)
			if err != nil {
				return err
			}

			// See which state machine field has chagned, and
			// if its a valid transition
			for _, f := range smFields {
				if post, ok := objMap[f]; ok {
					if post.(string) != smPreValues[f] {
						sm := getStateMachine(coll, f)
						if sm == nil {
							// TODO: Error
							return fmt.Errorf("no state machine found %s.%s", coll, f)
						} else if !sm.CanMove(smPreValues[f], post.(string)) {
							return fmt.Errorf("invalid state transition (%s) from %s to %s", f, smPreValues[f], post)
						}
					}
				}
			}
		}

		return nil
	}

	// If there is already a context available
	// then run it under it.
	// Otherwise, start a new transaction.
	if len(sessCtx) > 0 {
		err = fn(sessCtx[0])
	} else {
		err = mc.Transactionally(fn)
	}

	if err != nil {
		if manyErrs != nil && len(manyErrs) > 0 && manyErrs[0] == err {
			return manyErrs
		} else {
			return []ErrorPlus{{Message: err.Error()}}
		}
	}

	return []ErrorPlus{}
}
