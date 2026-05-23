package caco3

import (
	"encoding/json"
	"os"
)

func readJSONFile(file string, obj any) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, obj)
}

func writeJSONFile(file string, obj any) error {
	bs, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return os.WriteFile(file, bs, 0644)
}
