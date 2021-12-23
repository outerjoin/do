package do

import "fmt"

type ErrorPlus struct {
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

func (e ErrorPlus) Error() string {
	if e.Source == "" {
		return e.Message
	}
	return fmt.Sprintf("[%s] %s", e.Source, e.Message)
}
