package do

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rightjoin/rutl/conv"
)

/*
	FIELD VALIDATION

	insert: yes/no/[opt]
	update: yes/no/[opt]
	trim:   [yes]/no
	auto:
		prefix:
		uuid | alphanum(12)
	verify:
		email:
		rex(...)
		enum(abc|def|ghi)
	[Recursive: ??]
*/

func correctInitalState(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// Initial state must open be checked during INSERTS only
	if action != DB_INSERT {
		return errs
	}

	mt := TypeOf(modelType)
	mt = TypeDereference(mt)

	wc := WalkConfig{"json"}

	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fkey := wc.FieldKey(ft)
		fsm, _ := ParseType(ft.Tag.Get("state_machine"), reflect.TypeOf(false))
		if fsm.(bool) {
			// force string
			if data.HasKey(fkey) {
				data[fkey] = fmt.Sprint(data[fkey])
			}

			sm := getStateMachine(MongoCollectionName(modelType), fkey)
			if sm == nil {
				errs = append(errs, ErrorPlus{
					Message: "no state machine found",
					Source:  fkey,
				})
			} else {
				if !data.HasKey(fkey) {
					data[fkey] = sm.DefaultState
				} else {
					if !sm.CanStartWith(data[fkey].(string)) {
						errs = append(errs, ErrorPlus{
							Message: fmt.Sprintf("given state (%s) is not a valid start state", data[fkey]),
							Source:  fkey,
						})
					}
				}
			}
		}
	}
	return errs
}

func verifyInputs(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// Input validations as defined in 'verify' tag
	verify := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		if (action == DB_INSERT || action == DB_UPDATE) && data.HasKey(fname) {
			checks := getFieldTests(fld)
			for _, check := range checks {
				if success, message := check.Verify(fld.Type, data[fname]); !success {
					errs = append(errs, ErrorPlus{
						Message: message,
						Source:  strings.Join(keys, "."),
					})
				}
			}
		}
		return nil
	}
	StructWalk(modelType, WalkConfig{"json"}, data, verify)
	return errs
}

func convertFieldType(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// Convert types of fields from STRING to appropriate type
	// as it is specified in the Struct
	convert := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		inp, found := data[fname]
		inpStr, isStr := inp.(string)
		expType := fld.Type.String()

		if found && (action == DB_INSERT || action == DB_UPDATE) && isStr && expType != "string" {
			val, err := ParseType(inpStr, TypeDereference(fld.Type))
			if err == nil {
				data[fname] = val
			} else {
				errs = append(errs, ErrorPlus{
					Message: err.Error(),
					Source:  strings.Join(keys, "."),
				})
			}
		}

		return nil
	}

	StructWalk(modelType, WalkConfig{"json"}, data, convert)
	return errs
}

func populateTimedFields(modelType interface{}, action int, data Map) []ErrorPlus {

	errs := []ErrorPlus{}
	isMongoEntity := TypeComposedOf(modelType, MongoEntity{})

	// Manage timestamp fields (inserted_at / updated_at)
	// during insert / update of records - do this for only
	// MongoEntitys for now

	// setTimed := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
	if isMongoEntity && TypeComposedOf(modelType, Timed{}) {
		now := time.Now()
		switch action {
		case DB_INSERT:
			data["created_at"] = now
			data["updated_at"] = now
		case DB_UPDATE:
			data["updated_at"] = now
		}
	}

	// 	return nil
	// }
	// StructWalk(modelType, WalkConfig{"json"}, data, setTimed)
	return errs
}

func populateAutoFields(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}
	isMongoEntity := TypeComposedOf(modelType, MongoEntity{})

	// Set fields marked auto - to give them appropriate value upon insertion
	setAuto := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		if action == DB_INSERT && !data.HasKey(fname) && fld.Tag.Get("auto") != "" {
			auto := parseAutoField(fld)
			if auto != nil {
				if isMongoEntity && fname == "id" && fld.Tag.Get("bson") != "" {
					// give preference to bson
					data[fld.Tag.Get("bson")] = auto.Generate()
				} else {
					data[fname] = auto.Generate()
				}
			}
		}
		return nil
	}
	StructWalk(modelType, WalkConfig{"json"}, data, setAuto)
	return errs
}

func trimFields(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// Trim any input strings fields, unless markeed no (trim=no)
	trim := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		if (action == DB_INSERT || action == DB_UPDATE) && data.HasKey(fname) && fld.Tag.Get("trim") != "no" {
			str, isString := data[fname].(string)
			if isString {
				data[fname] = strings.TrimSpace(str)
			}
		}
		return nil
	}
	StructWalk(modelType, WalkConfig{"json"}, data, trim)
	return errs
}

func provideDefualts(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// During inserts, if input fields are not provided and a default value is provided
	// in the field tags then do use it
	setDefaults := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]
		defStr := fld.Tag.Get("default")
		if action == DB_INSERT && !data.HasKey(fname) && defStr != "" && fld.Tag.Get("insert") != "no" {
			data[fname] = defStr
		}
		return nil
	}
	StructWalk(modelType, WalkConfig{"json"}, data, setDefaults)
	return errs
}

func checkInsertableUpdatableDataInMap(modelType interface{}, action int, data Map) []ErrorPlus {
	errs := []ErrorPlus{}

	// Do validations for those fields wherein input fields are extra or
	// input fields are expected but missing
	checkInsertUpdate := func(fld reflect.StructField, data Map, keys ...string) []ErrorPlus {
		fname := keys[len(keys)-1]

		switch action {
		case DB_INSERT:
			if data.HasKey(fname) && fld.Tag.Get("insert") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon insertion", fname, data.GetOr(fname, nil))
				errs = append(errs, ErrorPlus{issue, strings.Join(keys, ".")})
			}
			if !data.HasKey(fname) && fld.Tag.Get("insert") == "yes" {
				issue := fmt.Sprintf("field '%s' needs a value upon insertion", fname)
				errs = append(errs, ErrorPlus{issue, strings.Join(keys, ".")})
			}
		case DB_UPDATE:
			if data.HasKey(fname) && fld.Tag.Get("update") == "no" {
				issue := fmt.Sprintf("field '%s' cannot be given a value (%v) upon updation", fname, data.GetOr(fname, nil))
				errs = append(errs, ErrorPlus{issue, strings.Join(keys, ".")})
			}
		}
		return nil
	}
	StructWalk(modelType, WalkConfig{"json"}, data, checkInsertUpdate)
	return errs
}

func customInsertUpdateChecks(modelType interface{}, action int, data Map) []ErrorPlus {

	errs := []ErrorPlus{}

	// Before Save Check
	if savable, ok := modelType.(DBInsertableUpdatable); ok {
		errs = append(errs, savable.BeforeSave(modelType, data)...)
	}

	// Before Insert Check
	if action == DB_INSERT {
		if ins, ok := modelType.(DBInsertable); ok {
			errs = append(errs, ins.BeforeInsert(modelType, data)...)
		}
	}

	// Before Update Check
	if action == DB_UPDATE {
		if upd, ok := modelType.(DBUpdatable); ok {
			errs = append(errs, upd.BeforeUpdate(modelType, data)...)
		}
	}

	// TODO:
	// Do recursively on nested struct fields

	return errs
}

func ModelValidateInputs(modelType interface{}, action int, data Map) []ErrorPlus {

	errs := []ErrorPlus{}
	isMongoEntity := TypeComposedOf(modelType, MongoEntity{})

	// Any form keys of nature abc.def get properly
	// exapnded into nested maps
	// abc.def : k => abc : map[def] : k
	{
		newMap := data.Unlevel()

		// Copy new map into data
		for k := range data {
			delete(data, k)
		}
		for k := range newMap {
			data[k] = newMap[k]
		}
	}

	if isMongoEntity {
		errs = append(errs, checkInsertableUpdatableDataInMap(modelType, action, data)...)
		errs = append(errs, provideDefualts(modelType, action, data)...)
		errs = append(errs, trimFields(modelType, action, data)...)
		errs = append(errs, populateTimedFields(modelType, action, data)...)
		errs = append(errs, populateAutoFields(modelType, action, data)...)
		errs = append(errs, correctInitalState(modelType, action, data)...)
		{
			// field conversion from str to int/string/etc wherever appropriate
			errs = append(errs, convertFieldType(modelType, action, data)...)
		}
		// Post field type conversion - do verification and custom checks
		errs = append(errs, verifyInputs(modelType, action, data)...)
		errs = append(errs, customInsertUpdateChecks(modelType, action, data)...)
	}

	return errs
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type fieldTest struct {
	Test   string
	Option string
}

func (ft fieldTest) Verify(t reflect.Type, v interface{}) (bool, string) {

	switch t.String() {
	case "string":
		vstr := fmt.Sprintf("%s", v)
		switch ft.Test {
		case "email":
			if govalidator.IsEmail(vstr) {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s is not a valid email", vstr)
			}
		case "rex":
			reg, err := regexp.Compile(ft.Option)
			if err != nil {
				return false, fmt.Sprintf("%s is not a valid regular expression", ft.Option)
			}
			if reg.MatchString(vstr) {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s does not match the regular expression", vstr)
			}
		case "enum":
			if strings.Contains(ft.Option, "|"+vstr+"|") {
				return true, ""
			} else {
				return false, fmt.Sprintf("%s must be one of predefined set", vstr)
			}
		}
	}

	return false, "validation not supported: " + ft.Test
}

func getFieldTests(f reflect.StructField) (fv []fieldTest) {
	fv = []fieldTest{}

	verify := f.Tag.Get("verify")
	if verify == "" {
		return
	}

	tasks := strings.Split(verify, ";")
	for _, task := range tasks {
		f := parseFieldTest(task)
		if f != nil {
			fv = append(fv, *f)
		}
	}
	return
}

func parseFieldTest(input string) *fieldTest {
	fv := fieldTest{}

	i := strings.Index(input, "(")
	if i == -1 {
		fv.Test = input
		return &fv
	}

	j := strings.LastIndex(input, ")")
	fv.Test = input[0:i]
	fv.Option = input[i+1 : j]

	// If enum then set | at beginging and
	// end of options
	if fv.Test == "enum" {
		fv.Option = strings.TrimSpace(fv.Option)
		if !strings.HasPrefix(fv.Option, "|") {
			fv.Option = "|" + fv.Option
		}
		if !strings.HasSuffix(fv.Option, "|") {
			fv.Option = fv.Option + "|"
		}
	}

	return &fv
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

// auto:"prefix:p-;uuid|alphanum(16)"
type autoField struct {
	Method string // uuid | alphanum(length)?
	Length int    // 0

	Prefix string // optional
}

func (af *autoField) Generate() string {
	val := af.Prefix
	switch af.Method {
	case "uuid":
		val += NewUUID()
	case "alphanum":
		val += NewAlhpaNum(af.Length)
	}

	return val
}

func parseAutoField(f reflect.StructField) *autoField {

	input := f.Tag.Get("auto")
	if input == "" {
		return nil
	}

	// Split by ;
	parts := strings.Split(input, ";")
	af := autoField{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "prefix:") {
			af.Prefix = part[7:]
		} else if part == "uuid" {
			af.Method = part
		} else if strings.HasPrefix(part, "alphanum(") {
			af.Method = "alphanum"
			count := part[9 : len(part)-1]
			af.Length = conv.IntOr(count, 16)
		} else {
			// TODO: log unsupported methods "panic"
			return nil
		}
	}

	return &af
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func ModelValidateObject(object interface{}) []ErrorPlus {
	errs := []ErrorPlus{}

	if serialize, ok := object.(DBSerialize); ok {
		errs = append(errs, serialize.AfterSave(object)...)
	}

	// TODO:
	// recursively execute DBSerialize.AfterSave on
	// all nested structs of type DBSerialize

	return errs
}
