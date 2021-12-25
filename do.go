package do

import (
	"time"

	"github.com/brianvoe/gofakeit"
)

func init() {
	//
	gofakeit.Seed(time.Now().UTC().UnixNano())
}
