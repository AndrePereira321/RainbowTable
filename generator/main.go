package main

import (
	"RainbowTable/common"
	"crypto/sha256"
	"errors"
	"strings"
)

func getEncoder(hashAlgorithm string) (Encoder, error) {
	ltType := strings.ToLower(hashAlgorithm)
	switch ltType {
	case "sha-256":
		return &Sha256Encoder{Hash: sha256.New()}, nil
	}
	return nil, errors.New("unsupported hash algorithm: " + hashAlgorithm)
}

func getRainbowGenerator() (*RainbowGenerator, error) {
	filePath, err := common.GetConfigFilePath()
	if err != nil {
		return nil, err
	}
	config, err := common.ReadConfig(filePath)
	if err != nil {
		return nil, err
	}

	return &RainbowGenerator{
		Config: config,
	}, nil
}

func main() {
	generator, err := getRainbowGenerator()
	if err != nil {
		panic(err)
	}
	err = generator.GenerateTable()
	if err != nil {
		panic(err)
	}
}
