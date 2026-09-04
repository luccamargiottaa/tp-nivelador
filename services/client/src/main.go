package main

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

const base = 10
const agencyIdBitSize = 8
const batchSizeBitSize = 8

func loadConfig() (client.Config, error) {
	agencyId := os.Getenv("AGENCY_ID")

	if agencyId == "" {
		return client.Config{}, errors.New("AGENCY_ID environment variable is required")
	}
	agencyIdNum64, err := strconv.ParseUint(agencyId, base, agencyIdBitSize)

	if err != nil {
		return client.Config{}, errors.New("AGENCY_ID environment variable should be a number")
	}
	agencyIdNum8 := uint8(agencyIdNum64)

	serverHost := os.Getenv("SERVER_HOST")

	if serverHost == "" {
		return client.Config{}, errors.New("SERVER_HOST environment variable is required")
	}
	serverPort := os.Getenv("SERVER_PORT")

	if serverPort == "" {
		return client.Config{}, errors.New("SERVER_PORT environment variable is required")
	}
	inputFile := os.Getenv("INPUT_FILE")

	if inputFile == "" {
		return client.Config{}, errors.New("INPUT_FILE environment variable is required")
	}
	outputFile := os.Getenv("OUTPUT_FILE")

	if outputFile == "" {
		return client.Config{}, errors.New("OUTPUT_FILE environment variable is required")
	}
	batchSize := os.Getenv("BATCH_SIZE")

	if batchSize == "" {
		return client.Config{}, errors.New("BATCH_SIZE environment variable is required")
	}
	batchSizeIdNum64, err := strconv.ParseUint(batchSize, base, batchSizeBitSize)

	if err != nil {
		return client.Config{}, errors.New("BATCH_SIZE environment variable should be a number")
	}
	batchSizeIdNum8 := uint8(batchSizeIdNum64)

	return client.Config{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   agencyIdNum8,
		InputFile:  inputFile,
		OutputFile: outputFile,
		BatchSize:  batchSizeIdNum8,
	}, nil
}

func listenToSigterm(sigCh chan os.Signal, doneCh chan bool, client *client.Client, shutdownDone *atomic.Bool) {
	signal.Notify(sigCh, syscall.SIGTERM)

	go func() {
		defer close(doneCh)

		if _, ok := <-sigCh; !ok {
			return
		}
		shutdownDone.Store(true)
		err := client.Conn.Close()

		if err != nil {
			logger.Error("shutdown", logger.Fail, "err", err)
		}
	}()
}

func closeChannel(sigCh chan os.Signal, doneCh chan bool) {
	signal.Stop(sigCh)
	close(sigCh)
	<-doneCh
}

func run() int {
	config, err := loadConfig()

	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}
	newClient, err := client.NewClient(config)

	if err != nil {
		logger.Error("client-new", logger.Fail, "err", err)
		return 1
	}
	sigCh := make(chan os.Signal, 1)
	doneCh := make(chan bool, 1)
	var shutdownDone atomic.Bool

	listenToSigterm(sigCh, doneCh, newClient, &shutdownDone)
	defer closeChannel(sigCh, doneCh)

	if err := newClient.Run(); err != nil {
		if shutdownDone.Load() {
			return 0
		}
		logger.Error("client-run", logger.Fail, "err", *err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
