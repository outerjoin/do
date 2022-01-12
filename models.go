package do

import (
	"fmt"
	"reflect"
	"time"
)

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
	MongoEntity

	ID           string                 `bson:"_id" json:"id" insert:"no" auto:"prefix:sm-;alphanum(8)"`
	Entity       string                 `bson:"entity" json:"entity"`
	Field        string                 `bson:"field" json:"field"`
	AllStates    []string               `bson:"all_states" json:"all_states"`
	StartStates  []string               `bson:"start_states" json:"start_states"`
	DefaultState string                 `bson:"default_state" json:"default_state"`
	Transitions  []StateMachineMovement `bson:"transitions" json:"transitions"`

	// Behaviours
	Timed `bson:"inline"`
}

func (sm StateMachine) AfterSave(object interface{}) []ErrorPlus {
	output := []ErrorPlus{}

	toValidate := object.(*StateMachine)

	// Default state must be one of Start states
	found := false
	for _, s := range toValidate.StartStates {
		if s == toValidate.DefaultState {
			found = true
			break
		}
	}
	if !found {
		output = append(output, ErrorPlus{Message: "default state must be one of start states"})
	}

	// Start states must be part of All states
	for _, start := range toValidate.StartStates {
		found = false
		for _, s := range toValidate.AllStates {
			if s == start {
				found = true
				break
			}
		}
		if !found {
			output = append(output, ErrorPlus{Message: fmt.Sprintf("start state (%s) must be one of all states", start)})
		}
	}

	return output
}

func (sm StateMachine) CanStartWith(state string) bool {
	if sm.StartStates != nil {
		for _, start := range sm.StartStates {
			if start == state {
				return true
			}
		}
	}
	return false
}

func (sm StateMachine) CanMove(from, to string) bool {
	if sm.Transitions != nil {
		for _, t := range sm.Transitions {
			if t.From == from && t.To == to {
				return true
			}
		}
	}

	return false
}

func (sm StateMachine) GetStateMachineFieldNames(obj interface{}) []string {
	output := []string{}

	mt := TypeOf(obj)
	mt = TypeDereference(mt)

	wc := WalkConfig{"json"}
	booltype := reflect.TypeOf(false)

	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fkey := wc.FieldKey(ft)
		fsm, _ := ParseType(ft.Tag.Get("state_machine"), booltype)
		if fsm.(bool) {
			output = append(output, fkey)
		}
	}

	return output
}

type StateMachineMovement struct {
	From string `bson:"from" json:"from"`
	To   string `bson:"to" json:"to"`
}

var allMachines []StateMachine = nil

func getStateMachine(entity, field string) *StateMachine {

	read := StateMachine{}
	records := []StateMachine{}
	var ref interface{} = &records

	if allMachines == nil {
		// Mongo connection
		mo := NewMongoConnect()
		defer mo.CloseClient()

		// Fetch
		_, err := mo.Query(read, ref)
		if err != nil {
			// TODO:
			// log error
			return nil
		}

		allMachines = records
	}

	for _, sm := range allMachines {
		if sm.Entity == entity && sm.Field == field {
			return &sm
		}
	}

	return nil
}
