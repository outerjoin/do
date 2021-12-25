package do

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/brianvoe/gofakeit"
)

var lookForPhony = regexp.MustCompile(`{{[a-zA-Z:0-9.]+}}`)

type Phony struct {
	jsonStr string
}

func NewPhony(text string) *Phony {
	return &Phony{
		jsonStr: text,
	}
}

func NewPhonyFromFile(filepath string) *Phony {
	b, err := FileContent(filepath)
	t := ""
	if err != nil {
		t = err.Error()
	} else {
		t = string(b)
	}
	return &Phony{
		jsonStr: t,
	}
}

func (d *Phony) String() string {

	return lookForPhony.ReplaceAllStringFunc(d.jsonStr, func(inp string) string {

		toReplace := inp[2 : len(inp)-2]
		vals := strings.Split(toReplace, ":")

		switch vals[0] {
		case "name":
			return gofakeit.Name()
		case "email":
			return gofakeit.Email()
		case "color":
			return gofakeit.Color()
		case "company":
			return gofakeit.Company()
		case "city":
			return gofakeit.City()
		case "street":
			return gofakeit.Street()
		case "state":
			return gofakeit.State()
		case "country":
			return gofakeit.Country()
		case "zip":
			return gofakeit.Zip()
		case "domain":
			return gofakeit.DomainName()
		case "url":
			return gofakeit.URL()
		case "gender":
			return gofakeit.Gender()
		case "phone":
			return gofakeit.Phone()
		case "date":
			return gofakeit.Date().Format("2006-01-02")
		case "datetime":
			return gofakeit.Date().Format("2006-01-02 15:04:05")
		case "weekday":
			return gofakeit.WeekDay()
		case "latitude":
			return fmt.Sprintf("%f", gofakeit.Latitude())
		case "longitude":
			return fmt.Sprintf("%f", gofakeit.Longitude())
		case "enum":
			return vals[1+rand.Intn(len(vals)-1)]
		case "price", "decimal":
			min := 0.0
			max := 1000.0
			if len(vals) == 2 {
				min = ParseFloat64Or(vals[1], 0)
				max = min + 1000
			} else if len(vals) == 3 {
				min = ParseFloat64Or(vals[1], 0)
				max = ParseFloat64Or(vals[2], min+1000)
			}
			return fmt.Sprintf("%0.2f", gofakeit.Price(min, max))
		case "number":
			min := 0
			max := 1000
			if len(vals) == 2 {
				min = ParseIntOr(vals[1], 0)
				max = min + 1000
			} else if len(vals) == 3 {
				min = ParseIntOr(vals[1], 0)
				max = ParseIntOr(vals[2], min+1000)
			}
			return fmt.Sprintf("%d", gofakeit.Number(min, max))
		}
		return inp
	})
}

func (d *Phony) Map() map[string]interface{} {

	var m map[string]interface{}
	err := json.Unmarshal([]byte(d.String()), &m)
	if err != nil {
		// TODO: log error
		return nil
	}
	return m
}

func (d *Phony) Array(len int) []map[string]interface{} {

	array := make([]map[string]interface{}, len)
	for i := 0; i < len; i++ {
		array[i] = d.Map()
	}
	return array
}
