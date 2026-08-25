package packets

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
)

const MaxNameLength = 50
const DocumentSize = 8
const DateSize = 10
const NumberSize = 4
const BetSize = MaxNameLength*2 + DocumentSize + DateSize + NumberSize

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
	copy(dest, bet.FirstName)
	copy(dest[MaxNameLength:], bet.LastName)
	copy(dest[MaxNameLength*2:], bet.Document)
	copy(dest[MaxNameLength*2+DocumentSize:], bet.BirthDate)

	binary.BigEndian.PutUint32(dest[MaxNameLength*2+DocumentSize+DateSize:], bet.Number)

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
	firstName := string(bytes[:MaxNameLength])
	lastName := string(bytes[MaxNameLength : MaxNameLength*2])
	document := string(bytes[MaxNameLength*2 : MaxNameLength*2+DocumentSize])
	birthDate := string(bytes[MaxNameLength*2+DocumentSize : MaxNameLength*2+DocumentSize+DateSize])

	number := binary.BigEndian.Uint32(bytes[MaxNameLength*2+DocumentSize+DateSize:])

	return &bet.Bet{FirstName: firstName, LastName: lastName, Document: document, BirthDate: birthDate, Number: number}, nil
}
