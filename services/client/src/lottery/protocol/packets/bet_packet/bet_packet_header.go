package bet_packet

import (
	"errors"
	"math"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

const MaxNameLength = math.MaxUint8

const firstNameLengthSize = 1
const lastNameLengthSize = 1
const BetPacketHeaderSize = firstNameLengthSize + lastNameLengthSize

const firstNameLengthIndex = 0
const lastNameLengthIndex = 1

type BetPacketHeader struct {
	FirstNameLength uint8
	LastNameLength  uint8
}

func NewBetPacketHeader(bet lottery.Bet) (*BetPacketHeader, error) {
	if len(bet.FirstName) > MaxNameLength {
		return nil, errors.New("first name is too long")
	}
	if len(bet.LastName) > MaxNameLength {
		return nil, errors.New("last name is too long")
	}
	return &BetPacketHeader{uint8(len(bet.FirstName)), uint8(len(bet.LastName))}, nil
}

func (header *BetPacketHeader) WriteToBytes(bytes []byte) {
	bytes[firstNameLengthIndex] = header.FirstNameLength
	bytes[lastNameLengthIndex] = header.LastNameLength
}

func BetPacketHeaderFromBytes(bytes []byte) (*BetPacketHeader, error) {
	if len(bytes) != BetPacketHeaderSize {
		return nil, errors.New("incorrect lottery packet header length")
	}
	return &BetPacketHeader{bytes[firstNameLengthIndex], bytes[lastNameLengthIndex]}, nil
}
