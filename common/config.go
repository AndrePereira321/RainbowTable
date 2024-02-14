package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const defaultBuffSize = 10_000

var ErrMissingConfigPath = errors.New("missing configuration file path as program argument")
var ErrInvalidConfigPath = errors.New("configuration file path program argument is invalid")

type RainbowConfig struct {
	Name           string  `json:"name"`
	HashAlgorithm  string  `json:"hashAlgorithm"`
	PasswordMin    int     `json:"passwordMin"`
	PasswordMax    int     `json:"passwordMax"`
	ChainLength    uint64  `json:"chainLength"`
	TableSize      uint64  `json:"tableSize"`
	ReduceSeed     string  `json:"reduceSeed"`
	WorkFolder     string  `json:"workFolder"`
	CoreQt         int     `json:"-"`
	CoreMultiplier float64 `json:"coreMultiplier"`
	BuffSize       uint64  `json:"buffSize"`
}

func (config *RainbowConfig) GetGeneratorFolder() string {
	return filepath.Join(config.WorkFolder, "generator")
}

func ReadConfig(filePath string) (*RainbowConfig, error) {
	buff, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading configuration file: %v", err)
	}

	var cfg RainbowConfig

	err = json.Unmarshal(buff, &cfg)
	if err != nil {

		return nil, fmt.Errorf("failed parsing json configuration file: %v", err)
	}

	cfg.CoreQt = runtime.NumCPU()
	if cfg.BuffSize == 0 {
		cfg.BuffSize = defaultBuffSize
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
