package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

const appendMode = os.O_WRONLY | os.O_CREATE | os.O_APPEND
const perm = 0644

const separator = ","
const betParts = 5

const firstNameIndex = 0
const lastNameIndex = 1
const documentIndex = 2
const birthdateIndex = 3
const numberIndex = 4

const base = 10
const bitSize = 32

type Storage struct {
	inputFile  *os.File
	outputFile *os.File
}

func NewStorage(inputPath string, outputPath string) (*Storage, error) {
	inputFile, err := os.Open(inputPath)

	if err != nil {
		return nil, err
	}
	outputFile, err := os.OpenFile(outputPath, appendMode, perm)

	if err != nil {
		closeFile(inputFile, &err)
		return nil, err
	}
	storage := &Storage{inputFile, outputFile}

	return storage, nil
}

func (storage *Storage) ReadBets(amount int) ([]lottery.Bet, error) {
	scanner := bufio.NewScanner(storage.inputFile)
	bets := make([]lottery.Bet, amount)

	for read := 0; read < amount && scanner.Scan(); read++ {
		line := scanner.Text()
		bet, err := betFromLine(line)

		if err != nil {
			return nil, err
		}
		bets[read] = *bet
	}
	return bets, scanner.Err()
}

func (storage *Storage) WriteBets(bets []lottery.Bet) (err *error) {
	writer := bufio.NewWriter(storage.outputFile)
	defer flushWriter(writer, err)

	for _, bet := range bets {
		line := betToLine(&bet)

		if _, writeErr := writer.WriteString(line); writeErr != nil {
			err = &writeErr
			return
		}
	}
	return
}

func (storage *Storage) Close() (err *error) {
	closeFile(storage.inputFile, err)
	closeFile(storage.outputFile, err)

	return
}

func betFromLine(line string) (bet *lottery.Bet, err error) {
	parts := strings.Split(line, separator)

	if len(parts) != betParts {
		return nil, errors.New("invalid lottery")
	}
	number, err := strconv.ParseUint(strings.TrimSpace(parts[numberIndex]), base, bitSize)

	if err != nil {
		return nil, err
	}
	document, err := strconv.ParseUint(strings.TrimSpace(parts[documentIndex]), base, bitSize)

	if err != nil {
		return nil, err
	}
	return &lottery.Bet{
		FirstName: strings.TrimSpace(parts[firstNameIndex]),
		LastName:  strings.TrimSpace(parts[lastNameIndex]),
		Document:  uint32(document),
		BirthDate: strings.TrimSpace(parts[birthdateIndex]),
		Number:    uint32(number),
	}, nil
}

func betToLine(bet *lottery.Bet) string {
	return fmt.Sprintf("%s,%s,%d,%s,%d", bet.FirstName, bet.LastName, bet.Document, bet.BirthDate, bet.Number)
}

func flushWriter(writer *bufio.Writer, err *error) {
	if flushError := writer.Flush(); flushError != nil {
		if err == nil {
			err = &flushError
		} else {
			*err = errors.Join(*err, flushError)
		}
	}
}

func closeFile(file *os.File, err *error) {
	if closeError := file.Close(); closeError != nil {
		if err == nil {
			err = &closeError
		} else {
			*err = errors.Join(*err, closeError)
		}
	}
}
