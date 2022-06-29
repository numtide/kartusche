package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func JSONEncoder(v interface{}) func() (io.Reader, error) {
	return func() (io.Reader, error) {
		bb := new(bytes.Buffer)
		err := json.NewEncoder(bb).Encode(v)
		if err != nil {
			return nil, fmt.Errorf("while encoding value: %w", err)
		}

		return bb, nil
	}
}
