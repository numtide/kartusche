package client

import (
	"encoding/json"
	"fmt"
	"io"
)

func JSONDecoder(v interface{}) func(r io.Reader) error {
	return func(r io.Reader) error {
		err := json.NewDecoder(r).Decode(v)
		if err != nil {
			return fmt.Errorf("while decoding JSON response: %w", err)
		}
		return nil
	}

}
