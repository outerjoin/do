package do

import "reflect"

func TypeComposedOf(modelOrType interface{}, parentModelOrType interface{}) bool {

	// Item type
	item := TypeOf(modelOrType)
	item = TypeDereference(item)

	// Parent type
	parent := TypeOf(parentModelOrType)
	parent = TypeDereference(parent)

	if parent.Kind() != reflect.Struct {
		// TODO: log error - parent's type must always be a struct
		return false
	}

	// find field with parent's exact name
	f, ok := item.FieldByName(parent.Name())
	if !ok {
		return false
	}

	if !f.Anonymous {
		return false
	}

	if !f.Type.ConvertibleTo(parent) {
		return false
	}

	return true
}

func TypeOf(modelOrType interface{}) reflect.Type {

	myType, isTypeAlready := modelOrType.(reflect.Type)
	if !isTypeAlready {
		myType = reflect.TypeOf(modelOrType)
	}

	return myType
}

func TypeDereference(mayBePtrType reflect.Type) reflect.Type {
	if mayBePtrType.Kind() == reflect.Ptr {
		return mayBePtrType.Elem()
	}
	return mayBePtrType
}

func TypeIsTime(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.Name() == "Time" && t.PkgPath() == "time"
}
