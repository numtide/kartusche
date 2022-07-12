package server

import (
	"encoding/json"
	"fmt"
)

func toJSON(v interface{}) []byte {
	d, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("while marshalling to JSON: %w", err))
	}
	return d
}
