package common

import (
	"encoding/json"
	"errors"
	"os"
)

var ErrMissingConfigPath = errors.New("missing configuration file path as program argument")
var ErrInvalidConfigPath = errors.New("configuration file path program argument is invalid")

type RainbowConfig struct {
	Name          string `json:"name"`
	HashAlgorithm string `json:"hashAlgorithm"`
	PasswordMin   int    `json:"passwordMin"`
	PasswordMax   int    `json:"passwordMax"`
	ChainLength   int    `json:"chainLength"`
	TableSize     int    `json:"tableSize"`
	ReduceSeed    string `json:"reduceSeed"`
	WorkFolder    string `json:"workFolder"`
}

func ReadConfig(filePath string) (*RainbowConfig, error) {
	buff, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg RainbowConfig

	err = json.Unmarshal(buff, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GetConfigFilePath() (string, error) {
	if len(os.Args) < 2 {
		return "", ErrMissingConfigPath
	}
	file := os.Args[1]
	if len(file) == 0 {
		return "", ErrInvalidConfigPath
	}
	return file, nil
}
