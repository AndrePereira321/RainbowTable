package main

import (
	"RainbowTable/common"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlInserter struct {
}

type MySqlInserterJob struct {
	Config   *common.RainbowConfig
	Con      *sql.Conn
	FilePath string
	ID       int
	WG       *sync.WaitGroup
}

var MissingMySqlConfigErr = errors.New("missing mysql configuration")

func (inserter *MySqlInserter) Insert(config *common.RainbowConfig) error {
	err := inserter.init(config)
	if err != nil {
		return err
	}

	client, err := inserter.getConnection(config)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	err = inserter.createStructure(client, config)
	if err != nil {
		return err
	}

	err = inserter.insertFiles(client, config)
	if err != nil {
		return err
	}

	return nil
}

func (inserter *MySqlInserter) init(config *common.RainbowConfig) error {
	if config.MySqlConfig == nil {
		return MissingMySqlConfigErr
	}
	return nil
}

func (inserter *MySqlInserter) getConnection(config *common.RainbowConfig) (*sql.DB, error) {
	dsn := config.MySqlConfig.Dsn()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	jobQt := config.GetJobQt()
	db.SetMaxIdleConns(jobQt / 2)
	db.SetMaxOpenConns(jobQt + 3)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging the database: %v", err)
	}

	return db, nil
}

func (inserter *MySqlInserter) insertFiles(client *sql.DB, config *common.RainbowConfig) error {
	generatorFolder := config.GetGeneratorFolder()
	files, err := inserter.getAllFiles(generatorFolder)
	if err != nil {
		return fmt.Errorf("error retrieving files paths: %v", err)
	}

	var wg sync.WaitGroup
	for i := range files {
		fileName := files[i]
		conn, connErr := client.Conn(context.TODO())
		if connErr != nil {
			return fmt.Errorf("error opening database connection: %v", connErr)
		}

		job := &MySqlInserterJob{
			Config:   config,
			Con:      conn,
			FilePath: fileName,
			ID:       i,
			WG:       &wg,
		}

		wg.Add(1)
		go job.insertFiles()
	}

	wg.Wait()

	return nil
}

func (job *MySqlInserterJob) insertFiles() {
	defer func() {
		//TODO This error should be handled?
		_ = job.Con.Close()
		job.WG.Done()
	}()

	return
}

func (inserter *MySqlInserter) createStructure(client *sql.DB, config *common.RainbowConfig) error {
	err := inserter.createDatabase(client, config.Name)
	if err != nil {
		return fmt.Errorf("error creating database: %v", err)
	}

	_, err = client.Exec("USE " + config.Name)
	if err != nil {
		return err
	}

	for i := config.PasswordMin; i <= config.PasswordMax; i++ {
		_, err = inserter.createTable(client, i)
		if err != nil {
			return fmt.Errorf("error creating table with length %d: %v", i, err)
		}
	}
	return nil
}

func (inserter *MySqlInserter) createDatabase(client *sql.DB, dbName string) error {
	rows, err := client.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE LOWER(SCHEMA_NAME) = LOWER(?)", dbName)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	if rows.Next() {
		return nil
	}

	_, err = client.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return err
	}
	return nil
}

func (inserter *MySqlInserter) createTable(client *sql.DB, colSize int) (string, error) {
	tableName := fmt.Sprintf("table_%d", colSize)

	rows, err := client.Query("SELECT table_name FROM information_schema.tables WHERE LOWER(table_name) = LOWER(?)", tableName)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = rows.Close()
	}()

	if rows.Next() {
		return tableName, nil
	}

	query := fmt.Sprintf("CREATE TABLE %s (last char(%d) PRIMARY KEY, first char(%d))", tableName, colSize, colSize)
	_, err = client.Exec(query)
	if err != nil {
		return "", err
	}

	return tableName, nil
}

func (inserter *MySqlInserter) getAllFiles(folderPath string) ([]string, error) {
	var absolutePaths []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			absolutePaths = append(absolutePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return absolutePaths, nil
}
