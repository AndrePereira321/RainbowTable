package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const defaultBuffSize = 10_000

var ErrMissingConfigPath = errors.New("missing configuration file path as program argument")
var ErrInvalidConfigPath = errors.New("configuration file path program argument is invalid")

type RainbowConfig struct {
	Name           string       `json:"name"`
	HashAlgorithm  string       `json:"hashAlgorithm"`
	PasswordMin    int          `json:"passwordMin"`
	PasswordMax    int          `json:"passwordMax"`
	ChainLength    uint64       `json:"chainLength"`
	TableSize      uint64       `json:"tableSize"`
	Separator      string       `json:"separator"`
	ReduceSeed     string       `json:"reduceSeed"`
	SeedScore      uint64       `json:"-"`
	WorkFolder     string       `json:"workFolder"`
	CoreQt         int          `json:"-"`
	CoreMultiplier float64      `json:"coreMultiplier"`
	BuffSize       uint64       `json:"buffSize"`
	Method         string       `json:"method"`
	MySqlConfig    *MySqlConfig `json:"mySqlConfig"`
}

type MySqlConfig struct {
	HostName string `json:"hostName"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (config *RainbowConfig) GetGeneratorFolder() string {
	return filepath.Join(config.WorkFolder, "generator")
}

func (config *RainbowConfig) GetJobQt() int {
	if config.CoreMultiplier > 0 {
		return int(float64(config.CoreQt) * config.CoreMultiplier)
	}
	return config.CoreQt
}

func (config *MySqlConfig) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/", config.User, config.Password, config.HostName)
}

func (config *MySqlConfig) GetTableName(colSize int) string {
	return fmt.Sprintf("table_%d", colSize)
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
	if len(cfg.Method) == 0 {
		cfg.Method = "mysql"
	}
	if len(cfg.Separator) == 0 {
		cfg.Separator = " "
	}

	seedScore := uint64(0)
	for i := 0; i < len(cfg.ReduceSeed) && cfg.ReduceSeed[i] != 0; i++ {
		seedScore += uint64(cfg.ReduceSeed[i])
		if i > 0 {
			seedScore ^= uint64(cfg.ReduceSeed[i])
		}
	}
	if seedScore == 0 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		seedScore = r.Uint64()
	}

	cfg.SeedScore = seedScore

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
