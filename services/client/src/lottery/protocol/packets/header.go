package packets

import (
	"encoding/binary"
	"errors"
)

const codeSize = 1
const betAmountSize = 1
const betSizeSize = 2
const agencyIdSize = 1
const HeaderSize = codeSize + betAmountSize + betSizeSize + agencyIdSize

const codeIndex = 0
const betAmountIndex = codeIndex + codeSize
const betSizeIndex = betAmountIndex + betAmountSize
const agencyIdIndex = betSizeIndex + betSizeSize

const AckCode = 0
const BetCode = 1
const EndCode = 2

type Header struct {
	Code      uint8
	BetAmount uint8
	BetSize   uint16
	AgencyId  uint8
}

func newHeader(code uint8, betAmount uint8, betSize uint16, agencyId uint8) Header {
	return Header{code, betAmount, betSize, agencyId}
}

func newNonBetsHeader(code uint8, agencyId uint8) Header {
	return newHeader(code, 0, 0, agencyId)
}

func NewAckHeader(agencyId uint8) Header {
	return newNonBetsHeader(AckCode, agencyId)
}

func NewBetHeader(betAmount uint8, betSize uint16, agencyId uint8) Header {
	return newHeader(BetCode, betAmount, betSize, agencyId)
}

func NewEndHeader(agencyId uint8) Header {
	return newNonBetsHeader(EndCode, agencyId)
}

func (header *Header) WriteToBytes(bytes []byte) {
	bytes[codeIndex] = header.Code
	bytes[betAmountIndex] = header.BetAmount

	binary.BigEndian.PutUint16(bytes[betSizeIndex:], header.BetSize)

	bytes[agencyIdIndex] = header.AgencyId
}

func HeaderFromBytes(bytes []byte) (*Header, error) {
	if len(bytes) != HeaderSize {
		return nil, errors.New("invalid header size")
	}
	code := bytes[codeIndex]
	betAmount := bytes[betAmountIndex]

	betSize := binary.BigEndian.Uint16(bytes[betSizeIndex:])

	agencyId := bytes[agencyIdIndex]

	if code == BetCode {
		if betAmount == 0 {
			return nil, errors.New("invalid bet amount")
		}
		if betSize == 0 {
			return nil, errors.New("invalid bet size")
		}
	} else if betAmount != 0 || betSize != 0 {
		return nil, errors.New("invalid header code")
	}
	return &Header{code, betAmount, betSize, agencyId}, nil
}
