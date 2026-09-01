from lottery import Bet
from ..constants import BYTE_ORDER
from .bet_packet_header import (
    BET_PACKET_HEADER_SIZE,
    BetPacketHeader,
)

_DOCUMENT_SIZE = 4
BIRTHDATE_SIZE = 10
_NUMBER_SIZE = 4
_BET_BASE_SIZE = _DOCUMENT_SIZE + BIRTHDATE_SIZE + _NUMBER_SIZE

_DOCUMENT_INDEX = BET_PACKET_HEADER_SIZE
_BIRTHDATE_INDEX = _DOCUMENT_INDEX + _DOCUMENT_SIZE
_NUMBER_INDEX = _BIRTHDATE_INDEX + BIRTHDATE_SIZE
_FIRST_NAME_INDEX = _NUMBER_INDEX + _NUMBER_SIZE


class BetPacket:
    def __init__(self, header: BetPacketHeader, bet: Bet):
        if len(bet.birthdate) != BIRTHDATE_SIZE:
            raise ValueError('incorrect birthdate size')

        self.header = header
        self.bet = bet

    @classmethod
    def from_bet(cls, bet: Bet) -> 'BetPacket':
        return cls(BetPacketHeader.from_bet(bet), bet)

    def write_to_bytes(self, bytes: bytearray) -> None:
        self.header.write_to_bytes(bytes)

        bytes.extend(
            self.bet.document.to_bytes(
                length=_DOCUMENT_SIZE,
                byteorder=BYTE_ORDER))

        bytes.extend(self.bet.birthdate.encode())

        bytes.extend(
            self.bet.number.to_bytes(
                length=_NUMBER_SIZE,
                byteorder=BYTE_ORDER))

        bytes.extend(self.bet.first_name.encode())
        bytes.extend(self.bet.last_name.encode())

    def size(self) -> int:
        return (
            BET_PACKET_HEADER_SIZE
            + _BET_BASE_SIZE
            + self.header.first_name_length
            + self.header.last_name_length
        )

    @classmethod
    def from_bytes(cls, bytes: bytes, agency_id: int) -> 'BetPacket':
        if len(bytes) < BET_PACKET_HEADER_SIZE + _BET_BASE_SIZE:
            raise ValueError('invalid bet packet length')

        header = BetPacketHeader.from_bytes(bytes[:BET_PACKET_HEADER_SIZE])

        document = int.from_bytes(
            bytes[_DOCUMENT_INDEX:_DOCUMENT_INDEX +
                  _DOCUMENT_SIZE], byteorder=BYTE_ORDER
        )
        birthdate = bytes[_BIRTHDATE_INDEX:_BIRTHDATE_INDEX +
                          BIRTHDATE_SIZE].decode()

        number = int.from_bytes(
            bytes[_NUMBER_INDEX:_NUMBER_INDEX +
                  _NUMBER_SIZE], byteorder=BYTE_ORDER
        )
        first_name_end_index = _FIRST_NAME_INDEX + header.first_name_length
        first_name = bytes[_FIRST_NAME_INDEX:first_name_end_index].decode()

        last_name = bytes[first_name_end_index: first_name_end_index +
                          header.last_name_length].decode()

        return cls(
            header,
            Bet(agency_id, first_name, last_name, document, birthdate, number),
        )
