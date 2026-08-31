package bet_packet

import (
	"encoding/binary"
	"errors"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
)

const documentSize = 4
const BirthdateSize = 10
const numberSize = 4
const betBaseSize = documentSize + BirthdateSize + numberSize

const documentIndex = BetPacketHeaderSize
const birthdateIndex = documentIndex + documentSize
const numberIndex = birthdateIndex + BirthdateSize
const firstNameIndex = numberIndex + numberSize

type BetPacket struct {
	header BetPacketHeader
	Bet    lottery.Bet
}

func NewBetPacket(bet lottery.Bet) (*BetPacket, error) {
	header, err := NewBetPacketHeader(bet)

	if err != nil {
		return nil, err
	}
	if len(bet.BirthDate) != BirthdateSize {
		return nil, errors.New("incorrect birthdate size")
	}
	return &BetPacket{*header, bet}, nil
}

func (packet *BetPacket) WriteToBytes(bytes []byte) {
	packet.header.WriteToBytes(bytes)

	binary.BigEndian.PutUint32(bytes[documentIndex:], packet.Bet.Document)
	copy(bytes[birthdateIndex:], packet.Bet.BirthDate)
	binary.BigEndian.PutUint32(bytes[numberIndex:], packet.Bet.Number)

	copy(bytes[firstNameIndex:], packet.Bet.FirstName)
	copy(bytes[firstNameIndex+packet.header.FirstNameLength:], packet.Bet.LastName)
}

func BetPacketFromBytes(bytes []byte) (*BetPacket, error) {
	header, err := BetPacketHeaderFromBytes(bytes[:BetPacketHeaderSize])

	if err != nil {
		return nil, err
	}
	document := binary.BigEndian.Uint32(bytes[documentIndex:])
	birthDate := string(bytes[birthdateIndex : birthdateIndex+BirthdateSize])
	number := binary.BigEndian.Uint32(bytes[numberIndex:])

	firstNameEndIndex := firstNameIndex + header.FirstNameLength
	firstName := string(bytes[firstNameIndex:firstNameEndIndex])
	lastName := string(bytes[firstNameEndIndex : firstNameEndIndex+header.LastNameLength])

	bet := lottery.Bet{FirstName: firstName, LastName: lastName, Document: document, BirthDate: birthDate, Number: number}

	return &BetPacket{*header, bet}, nil
}

func (packet *BetPacket) Size() uint16 {
	return uint16(BetPacketHeaderSize + betBaseSize + packet.header.FirstNameLength + packet.header.LastNameLength)
}
