package loom

import (
	"bytes"
	"encoding/json"
)

func prettyJSON(v any) (string, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
