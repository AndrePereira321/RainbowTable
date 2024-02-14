package main

import (
	"RainbowTable/common"
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"os"
	"path/filepath"
	"strings"
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
	_, err := job.Con.ExecContext(context.TODO(), "USE "+job.Config.Name)
	if err != nil {
		//TODO Something to do with this error?
		fmt.Println(err)
		return
	}

	file, err := os.Open(job.FilePath)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	var first, last string
	separator := fmt.Sprintf("%c", job.Config.Separator[0])
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//TODO BUlk insert?
		first, last = job.getLine(scanner.Text(), separator)
		if len(first) == 0 || len(last) == 0 || len(first) != len(last) {
			continue
		}
		err = job.insertValues(first, last)
		if err != nil {
			//TODO Something to do with this error?
			fmt.Println(err)
		}
	}

	if err = scanner.Err(); err != nil {
		//TODO Something to do with this error?
		return
	}

	return
}

func (job *MySqlInserterJob) getLine(input string, separator string) (string, string) {
	parts := strings.Split(input, separator)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func (job *MySqlInserterJob) insertValues(first, last string) error {
	tableName := job.Config.MySqlConfig.GetTableName(len(first))
	query := fmt.Sprintf("INSERT INTO %s (first, last) VALUES (?, ?)", tableName)

	stmt, err := job.Con.PrepareContext(context.TODO(), query)
	if err != nil {
		return fmt.Errorf("error preparing insert query: %v", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.Exec(first, last)
	if err != nil {
		if isDuplicateEntryError(err) {
			return nil
		}
		return fmt.Errorf("error executing insert query: %v", err)
	}

	return nil
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL error code for duplicate entry violation is 1062
		return mysqlErr.Number == 1062 || strings.Contains(mysqlErr.Message, "Duplicate entry")
	}
	return false
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
		_, err = inserter.createTable(client, config, i)
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

func (inserter *MySqlInserter) createTable(client *sql.DB, config *common.RainbowConfig, colSize int) (string, error) {
	tableName := config.MySqlConfig.GetTableName(colSize)
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
