package packets

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
)

const MaxNameLength = 50
const DocumentSize = 8
const BirthdateSize = 10
const numberSize = 4
const BetSize = MaxNameLength*2 + DocumentSize + BirthdateSize + numberSize

const firstNameIndex = 0
const lastNameIndex = MaxNameLength
const documentIndex = MaxNameLength * 2
const birthdateIndex = MaxNameLength*2 + DocumentSize
const numberIndex = MaxNameLength*2 + DocumentSize + BirthdateSize

const MaxBets = math.MaxUint8

type Packet struct {
	Header Header
	Bets   []bet.Bet
}

func newPacket(header Header, bets []bet.Bet) *Packet {
	return &Packet{header, bets}
}

func newHeaderPacket(header Header) *Packet {
	return newPacket(header, make([]bet.Bet, 0))
}

func NewAckPacket(agencyId uint8) *Packet {
	return newHeaderPacket(NewAckHeader(agencyId))
}

func NewBetPacket(agencyId uint8, bets []bet.Bet) (*Packet, error) {
	if len(bets) == 0 {
		return nil, errors.New("bets is empty")
	}
	if len(bets) > MaxBets {
		return nil, errors.New("too many bets")
	}
	return newPacket(NewBetHeader(uint8(len(bets)), agencyId), bets), nil
}

func NewEndPacket(agencyId uint8) *Packet {
	return newHeaderPacket(NewEndHeader(agencyId))
}

func (packet *Packet) writeBetToBytes(bet bet.Bet, dest []byte) error {
	if len(bet.FirstName) > MaxNameLength {
		return errors.New("first name is too long")
	}
	if len(bet.LastName) > MaxNameLength {
		return errors.New("last name is too long")
	}
	if len(bet.Document) != DocumentSize {
		return errors.New("incorrect document size")
	}
	if len(bet.BirthDate) != BirthdateSize {
		return errors.New("incorrect birthdate size")
	}
	copy(dest[firstNameIndex:], bet.FirstName)
	copy(dest[lastNameIndex:], bet.LastName)
	copy(dest[documentIndex:], bet.Document)
	copy(dest[birthdateIndex:], bet.BirthDate)

	binary.BigEndian.PutUint32(dest[numberIndex:], bet.Number)

	return nil
}

func (packet *Packet) ToBytes() ([]byte, error) {
	bytes := make([]byte, HeaderSize+BetSize*packet.Header.BetAmount)

	packet.Header.WriteToBytes(bytes)

	for i, b := range packet.Bets {
		err := packet.writeBetToBytes(b, bytes[HeaderSize+BetSize*i:])

		if err != nil {
			return nil, err
		}
	}
	return bytes, nil
}

func BetFromBytes(bytes []byte) (*bet.Bet, error) {
	if len(bytes) != BetSize {
		return nil, errors.New("invalid bet size")
	}
	firstName := string(bytes[firstNameIndex : firstNameIndex+MaxNameLength])
	lastName := string(bytes[lastNameIndex : lastNameIndex+MaxNameLength])
	document := string(bytes[documentIndex : documentIndex+DocumentSize])
	birthDate := string(bytes[birthdateIndex : birthdateIndex+BirthdateSize])

	number := binary.BigEndian.Uint32(bytes[numberIndex : numberIndex+numberSize])

	return &bet.Bet{FirstName: firstName, LastName: lastName, Document: document, BirthDate: birthDate, Number: number}, nil
}
