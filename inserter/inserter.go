package main

import "RainbowTable/common"

type Inserter interface {
	Insert(config *common.RainbowConfig) error
}
