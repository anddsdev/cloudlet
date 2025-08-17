package utils

import (
	"encoding/json"
	"io"
)

func WriteJSONData(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}
