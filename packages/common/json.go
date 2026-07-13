package common

import (
	"encoding/json"
	"io"
	"os"
)

func ToJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func ReadJSON(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func WriteJSON(path string, v interface{}) error {
	data, err := ToJSON(v)
	if err != nil {
		return err
	}
	return AtomicWrite(path, data, 0644)
}
