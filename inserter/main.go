package main

import (
	"RainbowTable/common"
	"errors"
)

func getInserter() (Inserter, *common.RainbowConfig, error) {
	filePath, err := common.GetConfigFilePath()
	if err != nil {
		return nil, nil, err
	}
	config, err := common.ReadConfig(filePath)
	if err != nil {
		return nil, nil, err
	}

	ltMethod := config.Method
	switch ltMethod {
	case "mysql":
		return &MySqlInserter{}, config, nil
	}
	return nil, nil, errors.New("unsupported method: " + config.Method)

}
func main() {
	inserter, config, err := getInserter()
	if err != nil {
		panic(err)
	}
	err = inserter.Insert(config)
	if err != nil {
		panic(err)
	}
}
