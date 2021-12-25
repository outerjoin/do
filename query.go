package do

import (
	"strings"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

func ExtractQueryBson(c echo.Context, modelType interface{}) (query bson.D, sort bson.M) {
	query = bson.D{}
	sort = bson.M{}

	params := c.QueryParams()
	for k, vals := range params {
		for _, v := range vals {
			if strings.HasPrefix(k, ":") {
				// :page or :chunk, so not part of sql quering
				continue
			}
			split := strings.Split(k, ":")
			key := split[0]
			op := ""
			if len(split) > 1 {
				op = split[1]
			}

			switch op {
			case "", "eq":
				query = append(query, bson.E{Key: key, Value: v})
			case "ord":
				sort[key] = ParseIntOr(v, 1)
			case "ne", "lt", "gt", "gte":
				dollarOp := "$" + op
				valType, found := StructGetFieldTypeByJsonKey(modelType, key)
				if !found {
					// TODO: log
					continue
				}
				val, err := ParseType(v, valType)
				if err != nil {
					// TODO: log
					continue
				}
				query = append(query, bson.E{Key: key, Value: bson.M{dollarOp: val}})
			}

		}
	}

	return
}
