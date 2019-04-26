package main

import (
	"bytes"
	"encoding/gob"
)

// GetBytes - Used for converting custom interface to Bytes
func GetBytes(key interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
