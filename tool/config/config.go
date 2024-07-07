package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const configFileName = `.relevan-see.json`

//go:embed default.json
var defaultConfig []byte

type ConfigData struct {
	Message string       `json:"skip-message"`
	Filters []FilterData `json:"filters"`
}
type FilterData struct {
	Name string          `json:"filter"`
	Data json.RawMessage `json:"data"`
}

func Load(root string) (ConfigData, error) {
	configLocation := filepath.Join(root, configFileName)
	jsonFile, openErr := os.Open(configLocation)
	var result ConfigData
	if openErr == nil {
		slog.Debug("Loading configuration from", slog.String("location", configLocation))
		defer jsonFile.Close()
		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal([]byte(byteValue), &result)
		return result, err
	} else {
		slog.Debug("Using default configuration")
		var err = json.Unmarshal(defaultConfig, &result)
		return result, err
	}
}

func DumpConfig(root string) {
	configLocation := filepath.Join(root, configFileName)
	fmt.Printf("Writing configuration to %s", configLocation)
	data := []byte(defaultConfig)
	os.WriteFile(configLocation, data, 0644)
}
