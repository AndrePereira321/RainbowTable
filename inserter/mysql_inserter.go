package main

import (
	"RainbowTable/common"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

import _ "github.com/go-sql-driver/mysql"

type MySqlInserter struct {
}

var MissingMySqlConfigErr = errors.New("missing mysql configuration")

func (inserter *MySqlInserter) Insert(config *common.RainbowConfig) error {
	err := inserter.init(config)
	if err != nil {
		return err
	}

	dsn := config.MySqlConfig.Dsn()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	con, err := db.Conn(context.TODO())
	if err != nil {
		return fmt.Errorf("error creating con: %v", err)
	}
	defer func() {
		_ = con.Close()
	}()

	fmt.Println("Connection: ")
	fmt.Println(con)
	return nil
}

func (inserter *MySqlInserter) init(config *common.RainbowConfig) error {
	if config.MySqlConfig == nil {
		return MissingMySqlConfigErr
	}
	return nil
}
