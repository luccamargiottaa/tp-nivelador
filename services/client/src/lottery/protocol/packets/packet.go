package packets

import (
	"errors"
	"math"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery/protocol/packets/bet_packet"
)

const MaxBets = math.MaxUint8

type Packet struct {
	header Header
	bets   []bet_packet.BetPacket
}

func newHeaderPacket(header Header) *Packet {
	bets := make([]bet_packet.BetPacket, 0)

	return &Packet{header, bets}
}

func NewAckPacket(agencyId uint8) *Packet {
	return newHeaderPacket(NewAckHeader(agencyId))
}

func NewBetPacket(agencyId uint8, bets []lottery.Bet) (*Packet, error) {
	betAmount := uint8(len(bets))

	if betAmount == 0 {
		return nil, errors.New("bets is empty")
	}
	if betAmount > MaxBets {
		return nil, errors.New("too many bets")
	}
	betPackets := make([]bet_packet.BetPacket, betAmount)

	var betSize uint16 = 0

	for i, bet := range bets {
		betPacket, err := bet_packet.NewBetPacket(bet)

		if err != nil {
			return nil, err
		}
		betPackets[i] = *betPacket
		betSize += betPacket.Size()
	}
	header := NewBetHeader(betAmount, betSize, agencyId)
	return &Packet{header, betPackets}, nil
}

func NewEndPacket(agencyId uint8) *Packet {
	return newHeaderPacket(NewEndHeader(agencyId))
}

func (packet *Packet) ToBytes() ([]byte, error) {
	bytes := make([]byte, HeaderSize+packet.header.BetSize)

	packet.header.WriteToBytes(bytes)

	var i uint16 = HeaderSize

	for _, bet := range packet.bets {
		bet.WriteToBytes(bytes[i:])
		i += bet.Size()
	}
	return bytes, nil
}

func AddBetsFromBytes(header *Header, bytes []byte, bets []lottery.Bet) error {
	var i uint16

	for i < uint16(header.BetAmount) {
		packet, err := bet_packet.BetPacketFromBytes(bytes[i:])

		if err != nil {
			return err
		}
		bets = append(bets, packet.ToBet())
		i += packet.Size()
	}
	return nil
}
