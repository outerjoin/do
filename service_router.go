package do

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/logrusorgru/aurora"
)

type POST struct{}
type PUT struct{}
type DELETE struct{}
type GET struct{}

var tGET = reflect.TypeOf(GET{})
var tPOST = reflect.TypeOf(POST{})
var tPUT = reflect.TypeOf(PUT{})
var tDELETE = reflect.TypeOf(DELETE{})

type Service struct {
}

type ServiceManager struct {
	echo *echo.Echo
}

func NewServiceManager(framework interface{}) ServiceManager {
	webEcho, isEcho := framework.(*echo.Echo)

	if !isEcho {
		// TODO: panic
	}

	return ServiceManager{
		echo: webEcho,
	}
}

var callCount = 0

func (e ServiceManager) Add(svcModel interface{}) {

	callCount++

	// Find underlying types and values
	svcType := reflect.TypeOf(svcModel)
	svcValue := reflect.ValueOf(svcModel)
	if svcType.Kind() == reflect.Ptr {
		svcType = svcType.Elem()
		svcValue = svcValue.Elem()
	}

	// If base Service is not implemented
	// then skip it
	base, found := svcType.FieldByName("Service")
	if !found || base.Type != reflect.TypeOf(Service{}) {
		return
	}

	// Loop all fields
	fieldCount := svcType.NumField()
	for i := 0; i < fieldCount; i++ {
		fld := svcType.Field(i)

		if fld.Name == "Service" {
			continue
		}

		// To find method name, first use the Tag
		methodName := fld.Tag.Get("invoke")
		if methodName == "" {
			methodName = strings.ToUpper(fld.Name[0:1]) + fld.Name[1:]
		}
		method := svcValue.MethodByName(methodName)
		if !method.IsValid() {
			continue
		}

		urlRoute := buildUrl(fld.Tag, base.Tag)
		handler := methodHandler(method)

		typeMatched := true
		switch fld.Type {
		case tGET:
			e.echo.GET(urlRoute, handler)
		case tPUT:
			e.echo.PUT(urlRoute, handler)
		case tPOST:
			e.echo.POST(urlRoute, handler)
		case tDELETE:
			e.echo.DELETE(urlRoute, handler)
		default:
			typeMatched = false
		}

		if typeMatched {
			httpMethod := fld.Type.Name()
			if callCount%2 == 0 {
				fmt.Println(spaces(8-len(httpMethod)), aurora.Cyan(httpMethod), urlRoute)
			} else {
				fmt.Println(spaces(8-len(httpMethod)), aurora.Magenta(httpMethod), urlRoute)
			}
		}
	}
}

func buildUrl(fieldTag, svcTag reflect.StructTag) string {
	prefix := svcTag.Get("route-prefix")
	route := fieldTag.Get("route")
	version := fieldTag.Get("version")

	url := route

	// Put prefix in front
	if prefix != "" {
		if url == "" {
			url = prefix
		} else {
			url = prefix + "/" + url
		}
	}

	// And version towards the end
	if version != "" {
		url = url + "/v" + version
	}

	return CleanFilePath(url)
}

func methodHandler(method reflect.Value) func(c echo.Context) error {
	return func(ec echo.Context) error {
		out := method.Call([]reflect.Value{reflect.ValueOf(ec)})
		outErr := out[0]
		if outErr.IsNil() {
			return nil
		} else {
			err, ok := outErr.Interface().(error)
			if ok {
				return err
			}
			return fmt.Errorf("expected error but found %v", outErr.Interface())
		}
	}
}

func spaces(count int) string {
	out := ""
	for i := 1; i <= count; i++ {
		out += " "
	}
	return out
}
