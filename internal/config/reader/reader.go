package reader

import (
	"encoding/json"
	"os"
)

func LoadJsonConfig[T any](filepath string) (T, error) {
	var config T

	// open file
	file, err := os.Open(filepath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	// decode JSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}
