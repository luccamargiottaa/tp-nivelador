package client

import (
	"errors"
	"net"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery/storage"
)

const connectionAttemptsMax = 3
const connectionAttemptsDelayMs = 200

const betsPerSend = 1

type Config struct {
	ServerHost string
	ServerPort string
	AgencyId   uint8
	InputFile  string
	OutputFile string
}

type Client struct {
	conn    net.Conn
	config  Config
	storage storage.Storage
}

func NewClient(config Config) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)

	if err != nil {
		return nil, err
	}
	stor, err := storage.NewStorage(config.InputFile, config.OutputFile)

	if err != nil {
		closeConnection(conn, &err)
		return nil, err
	}
	client := &Client{conn: conn, config: config, storage: *stor}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)

	for i := range connectionAttemptsMax {
		conn, err = net.Dial("tcp", host+":"+port)

		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(connectionAttemptsDelayMs * time.Millisecond)
			continue
		}
		logger.Info(action, logger.Success)
		break
	}
	return conn, err
}

func closeConnection(conn net.Conn, err *error) {
	if closeConnectionError := conn.Close(); closeConnectionError != nil {
		if err == nil {
			err = &closeConnectionError
		} else {
			*err = errors.Join(*err, closeConnectionError)
		}
	}
}

func closeStorage(storage storage.Storage, err *error) {
	if closeError := storage.Close(); closeError != nil {
		if err == nil {
			err = closeError
		} else {
			*err = errors.Join(*err, *closeError)
		}
	}
}

func (client *Client) Run() (err *error) {
	defer closeStorage(client.storage, err)
	defer closeConnection(client.conn, err)

	action := "send-bets"

	logger.Info(action, logger.InProgress)

	for {
		bets, readErr := client.storage.ReadBets(betsPerSend)

		if readErr != nil {
			err = &readErr
			return
		}
		if len(bets) == 0 {
			break
		}
		if sendErr := protocol.SendBets(client.conn, client.config.AgencyId, bets); sendErr != nil {
			err = &sendErr
			return
		}
	}
	logger.Info(action, logger.Success)

	action = "recv-winners"

	logger.Info(action, logger.InProgress)

	winners, recvErr := protocol.RecvWinners(client.conn, client.config.AgencyId)

	if recvErr != nil {
		err = &recvErr
		return
	}
	if writeErr := client.storage.WriteBets(winners); writeErr != nil {
		err = writeErr
		return
	}
	logger.Info(action, logger.Success)

	return
}
