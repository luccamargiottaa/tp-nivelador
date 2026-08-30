package bet_packet

import (
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
)

const documentBase = 10
const documentBitSize = 32

const documentSize = documentBitSize / 8
const BirthdateSize = 10
const numberSize = 4
const betBaseSize = documentSize + BirthdateSize + numberSize

const documentIndex = BetPacketHeaderSize
const birthdateIndex = documentIndex + documentSize
const numberIndex = birthdateIndex + BirthdateSize
const firstNameIndex = numberIndex + numberSize

type BetPacket struct {
	header    BetPacketHeader
	firstName string
	lastName  string
	document  uint32
	birthdate string
	number    uint32
}

func NewBetPacket(bet bet.Bet) (*BetPacket, error) {
	header, err := NewBetPacketHeader(bet)

	if err != nil {
		return nil, err
	}
	document64, err := strconv.ParseUint(bet.Document, documentBase, documentBitSize)

	if err != nil {
		return nil, err
	}
	document := uint32(document64)

	if len(bet.BirthDate) != BirthdateSize {
		return nil, errors.New("incorrect birthdate size")
	}
	return &BetPacket{*header, bet.FirstName, bet.LastName, document, bet.BirthDate, bet.Number}, nil
}

func (packet *BetPacket) WriteToBytes(bytes []byte) {
	packet.header.WriteToBytes(bytes)

	binary.BigEndian.PutUint32(bytes[documentIndex:], packet.document)
	copy(bytes[birthdateIndex:], packet.birthdate)
	binary.BigEndian.PutUint32(bytes[numberIndex:], packet.number)

	copy(bytes[firstNameIndex:], packet.firstName)
	copy(bytes[firstNameIndex+len(packet.firstName):], packet.lastName)
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

	return &BetPacket{*header, firstName, lastName, document, birthDate, number}, nil
}

func (packet *BetPacket) Size() uint16 {
	return uint16(BetPacketHeaderSize + betBaseSize + packet.header.FirstNameLength + packet.header.LastNameLength)
}

func (packet *BetPacket) ToBet() bet.Bet {
	document := strconv.FormatUint(uint64(packet.document), documentBase)

	return bet.Bet{FirstName: packet.firstName, LastName: packet.lastName, Document: document, BirthDate: packet.birthdate, Number: packet.number}
}
