package do

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/araddon/dateparse"
)

func ParseIntOr(str string, deflt int) int {
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return deflt
	}
	return int(i)
}

func ParseFloat32Or(str string, deflt float32) float32 {
	f, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return deflt
	}
	return float32(f)
}

func ParseFloat64Or(str string, deflt float64) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return deflt
	}
	return f
}

func ParseType(str string, t reflect.Type) (interface{}, error) {

	switch t.String() {
	case "string": // If its string just return it
		return str, nil
	case "int8":
		i, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return int8(i), nil
		}
		return nil, err
	case "int16":
		i, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return int16(i), nil
		}
		return nil, err
	case "int":
		i, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return int(i), nil
		}
		return nil, err
	case "int64":
		return strconv.ParseInt(str, 10, 64)
	case "uint8":
		i, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return uint8(i), nil
		}
		return nil, err
	case "uint16":
		i, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return uint16(i), nil
		}
		return nil, err
	case "uint":
		i, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return uint(i), nil
		}
		return nil, err
	case "uint64":
		return strconv.ParseUint(str, 10, 64)
	case "float32":
		f, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return float32(f), nil
		}
		return nil, err
	case "float64":
		return strconv.ParseFloat(str, 64)
	case "bool":
		return (str == "yes" || str == "true" || str == "1" || str == "Y" || str == "y"), nil
	case "time.Time":
		return dateparse.ParseAny(str)
	}
	return nil, errors.New("type not handled: " + t.String())
}
