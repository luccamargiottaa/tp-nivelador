package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
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

const numberBase = 10
const numberSize = 32

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

func (storage *Storage) ReadBets(amount int) ([]bet.Bet, error) {
	scanner := bufio.NewScanner(storage.inputFile)
	bets := make([]bet.Bet, amount)

	for read := 0; read < amount && scanner.Scan(); read++ {
		line := scanner.Text()
		b, err := betFromLine(line)

		if err != nil {
			return nil, err
		}
		bets[read] = *b
	}
	return bets, scanner.Err()
}

func (storage *Storage) WriteBets(bets []bet.Bet) (err *error) {
	writer := bufio.NewWriter(storage.outputFile)
	defer flushWriter(writer, err)

	for _, b := range bets {
		line := betToLine(&b)

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

func betFromLine(line string) (b *bet.Bet, err error) {
	parts := strings.Split(line, separator)

	if len(parts) != betParts {
		return nil, errors.New("invalid bet")
	}
	number, err := strconv.ParseUint(strings.TrimSpace(parts[numberIndex]), numberBase, numberSize)

	if err != nil {
		return nil, err
	}
	return &bet.Bet{
		FirstName: strings.TrimSpace(parts[firstNameIndex]),
		LastName:  strings.TrimSpace(parts[lastNameIndex]),
		Document:  strings.TrimSpace(parts[documentIndex]),
		BirthDate: strings.TrimSpace(parts[birthdateIndex]),
		Number:    uint32(number),
	}, nil
}

func betToLine(bet *bet.Bet) string {
	return fmt.Sprintf("%s,%s,%s,%s,%d", bet.FirstName, bet.LastName, bet.Document, bet.BirthDate, bet.Number)
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
