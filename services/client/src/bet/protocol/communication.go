package protocol

import (
	"errors"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/bet/protocol/packets"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

func sendPacket(socket io.ReadWriter, packet *packets.Packet) error {
	bytes, err := packet.ToBytes()

	if err != nil {
		return err
	}
	err = safe_socket.SendAll(socket, bytes)

	if err != nil {
		return err
	}
	return nil
}

func recvHeader(socket io.ReadWriter, agencyId uint8) (*packets.Header, error) {
	answer, err := safe_socket.RecvAll(socket, packets.HeaderSize)

	if err != nil {
		return nil, err
	}
	header, err := packets.HeaderFromBytes(answer)

	if err != nil {
		return nil, err
	}
	if header.AgencyId != agencyId {
		return nil, errors.New("unexpected agency id")
	}
	return header, nil
}

func expectAck(socket io.ReadWriter, agencyId uint8) error {
	header, err := recvHeader(socket, agencyId)

	if err != nil {
		return err
	}
	if header.Code != packets.AckCode {
		return errors.New("unexpected answer from server")
	}
	return nil
}

func SendBets(socket io.ReadWriter, agencyId uint8, bets []bet.Bet) error {
	betPacket, err := packets.NewBetPacket(agencyId, bets)

	if err != nil {
		return err
	}
	err = sendPacket(socket, betPacket)

	if err != nil {
		return err
	}
	return expectAck(socket, agencyId)
}

func RecvWinners(socket io.ReadWriter, agencyId uint8) ([]bet.Bet, error) {
	end := packets.NewEndPacket(agencyId)
	err := sendPacket(socket, end)

	if err != nil {
		return nil, err
	}
	err = expectAck(socket, agencyId)

	if err != nil {
		return nil, err
	}
	bets := make([]bet.Bet, 0)
	ack := packets.NewAckPacket(agencyId)

	for {
		header, err := recvHeader(socket, agencyId)

		if err != nil {
			return nil, err
		}
		if header.Code == packets.EndCode {
			err = sendPacket(socket, ack)

			if err != nil {
				return nil, err
			}
			break
		}
		if header.Code != packets.BetCode {
			return nil, errors.New("unexpected message from server")
		}
		bytes, err := safe_socket.RecvAll(socket, int(header.BetAmount*packets.BetSize))

		if err != nil {
			return nil, err
		}
		for i := 0; i < int(header.BetAmount); i++ {
			b, err := packets.BetFromBytes(bytes[i : i+1])

			if err != nil {
				return nil, err
			}
			bets = append(bets, *b)
		}
		err = sendPacket(socket, ack)

		if err != nil {
			return nil, err
		}
	}
	return bets, nil
}
