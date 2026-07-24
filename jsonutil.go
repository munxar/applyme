package main

import (
	"bytes"
	"encoding/json"
)

// marshalJSONIndent renders v as indented JSON with plain UTF-8 output.
// encoding/json's Marshal/MarshalIndent escape &, < and > to & etc. by
// default (a safeguard for embedding JSON in HTML), which mangles the raw
// data some job-ad sources send us. Disable it so output stays readable.
func marshalJSONIndent(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
