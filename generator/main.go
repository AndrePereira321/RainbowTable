package main

import (
	"RainbowTable/common"
	"fmt"
)

func GetRainbowGenerator() (*RainbowGenerator, error) {
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
	generator, err := GetRainbowGenerator()
	if err != nil {
		panic(err)
	}

	fmt.Println(generator.Config.Name)
	fmt.Println(generator.Config.HashAlgorithm)
}
