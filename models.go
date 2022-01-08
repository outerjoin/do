package do

import "time"

const (
	DB_SELECT = iota
	DB_INSERT
	DB_UPDATE
	DB_DELETE
)

type MongoEntity struct{}

type DBInsertable interface {
	BeforeInsert(modelOrType interface{}, data Map) []ErrorPlus
}

type DBUpdatable interface {
	BeforeUpdate(modelOrType interface{}, data Map) []ErrorPlus
}

type DBInsertableUpdatable interface {
	BeforeSave(modelOrType interface{}, data Map) []ErrorPlus
}

type DBSerialize interface {
	AfterSave(object interface{}) []ErrorPlus
}

type Timed struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at" index:"true"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at" index:"true"`
}

type StateMachine struct {
	ID           string `bson:"_id" json:"id" insert:"no" auto:"prefix:sm-;alphanum(8)"`
	Entity       string `bson:"entity" json:"entity"`
	Field        string `bson:"field" json:"field"`
	States       *[]string
	EntryStates  *[]string
	DefaultState *string
	Transitions  *[]StateMachineMovement

	// Behaviours
	Timed
}

type StateMachineMovement struct {
	From string `bson:"from" json:"from"`
	To   string `bson:"to" json:"to"`
}
