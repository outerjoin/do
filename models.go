package do

import "time"

const (
	DB_SELECT = iota
	DB_INSERT
	DB_UPDATE
	DB_DELETE
)

type MongoEntity struct{}

type Timed struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at" index:"true"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at" index:"true"`
}
