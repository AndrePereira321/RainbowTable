package main

import (
	"RainbowTable/common"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type RainbowGenerator struct {
	Config *common.RainbowConfig
}

type GeneratorJob struct {
	Config         *common.RainbowConfig
	ID             int
	WG             *sync.WaitGroup
	RowsToGenerate uint64
	Encoder        Encoder
	Rand           *rand.Rand
}

func (gen *RainbowGenerator) GenerateTable() error {
	var wg sync.WaitGroup

	err := gen.init()
	if err != nil {
		return err
	}

	jobQt := gen.Config.GetJobQt()
	rowsPerJob := gen.Config.TableSize / uint64(jobQt)

	for i := 0; i < jobQt; i++ {
		encoder, encoderErr := getEncoder(gen.Config.HashAlgorithm)
		if encoderErr != nil {
			return encoderErr
		}

		job := &GeneratorJob{
			Config:         gen.Config,
			WG:             &wg,
			RowsToGenerate: rowsPerJob,
			ID:             i,
			Encoder:        encoder,
			Rand:           rand.New(rand.NewSource(time.Now().UnixNano() * int64(i+1))),
		}

		wg.Add(1)
		go job.GenerateTable()

		time.Sleep(7 * time.Millisecond)
	}

	wg.Wait()

	return nil
}

func (job *GeneratorJob) GenerateTable() {
	defer job.WG.Done()

	var buffer bytes.Buffer
	buffSize := job.Config.BuffSize

	fileIndex := uint64(0)
	separator := job.Config.Separator[0]

	var first, last, hash []byte
	var pwLen int
	for i := uint64(0); i < job.RowsToGenerate; i++ {
		if i > 0 && i%buffSize == 0 {
			//TODO May we report that error? Writing in the file in another routine?
			_ = job.writeToFile(buffer.Bytes())
			buffer.Reset()
			fileIndex++
		}

		first = common.GenerateRandomString(job.Rand, job.Config.PasswordMin, job.Config.PasswordMax)
		pwLen = len(first)
		last = make([]byte, pwLen)
		copy(last, first)

		hash = job.Encoder.Encode(last, nil)
		for j := uint64(0); j < job.Config.ChainLength; j++ {
			common.ReduceHash(hash, j, job.Config.SeedScore, last)
			job.Encoder.Encode(last, hash)
		}

		buffer.Write(first)
		buffer.WriteByte(separator)
		buffer.Write(last)
		buffer.WriteRune('\n')
	}

	return
}

func (job *GeneratorJob) writeToFile(buff []byte) error {
	jobId := strconv.Itoa(job.ID)
	filePath := filepath.Join(job.Config.GetGeneratorFolder(), "generated_"+strconv.Itoa(job.ID)+".txt")

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening buffer file of for job "+jobId+":", err)
	}

	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(buff)
	if err != nil {
		return fmt.Errorf("error writing buffer in file of for job "+jobId+":", err)
	}

	return nil
}

func (gen *RainbowGenerator) init() error {
	return gen.createWorkingDirectory()
}

func (gen *RainbowGenerator) createWorkingDirectory() error {
	generatorFolder := gen.Config.GetGeneratorFolder()
	_, err := os.Stat(generatorFolder)

	if os.IsNotExist(err) {
		dirErr := os.Mkdir(generatorFolder, os.ModePerm)
		if dirErr != nil {
			return fmt.Errorf("failed to create generator directory: %v", dirErr)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to check generator folder existence: %v", err)
	}

	return nil
}
