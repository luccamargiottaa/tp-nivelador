package packets

import "errors"

const codeSize = 1
const betAmountSize = 1
const agencyIdSize = 1
const HeaderSize = codeSize + betAmountSize + agencyIdSize

const codeIndex = 0
const betAmountIndex = 1
const agencyIdIndex = 2

const AckCode = 0
const BetCode = 1
const EndCode = 2

type Header struct {
	Code      uint8
	BetAmount uint8
	AgencyId  uint8
}

func newHeader(code uint8, betAmount uint8, agencyId uint8) Header {
	return Header{code, betAmount, agencyId}
}

func newNonBetsHeader(code uint8, agencyId uint8) Header {
	return newHeader(code, 0, agencyId)
}

func NewAckHeader(agencyId uint8) Header {
	return newNonBetsHeader(AckCode, agencyId)
}

func NewBetHeader(betAmount uint8, agencyId uint8) Header {
	return newHeader(BetCode, betAmount, agencyId)
}

func NewEndHeader(agencyId uint8) Header {
	return newNonBetsHeader(EndCode, agencyId)
}

func (header *Header) WriteToBytes(dest []byte) {
	dest[codeIndex] = header.Code
	dest[betAmountIndex] = header.BetAmount
	dest[agencyIdIndex] = header.AgencyId
}

func HeaderFromBytes(header []byte) (*Header, error) {
	if len(header) != 3 {
		return nil, errors.New("invalid header length")
	}
	code := header[codeIndex]
	betAmount := header[betAmountIndex]
	agencyId := header[agencyIdIndex]

	if code != BetCode && betAmount != 0 {
		return nil, errors.New("invalid header code")
	}
	if betAmount == 0 {
		return nil, errors.New("invalid bet amount")
	}
	return &Header{code, betAmount, agencyId}, nil
}
