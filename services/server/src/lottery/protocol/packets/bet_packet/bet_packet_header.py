from lottery import Bet
from ..constants import BYTE_ORDER

MAX_NAME_LENGTH = 120

_FIRST_NAME_LENGTH_SIZE = 1
_LAST_NAME_LENGTH_SIZE = 1
BET_PACKET_HEADER_SIZE = _FIRST_NAME_LENGTH_SIZE + _LAST_NAME_LENGTH_SIZE

_FIRST_NAME_LENGTH_INDEX = 0
_LAST_NAME_LENGTH_INDEX = _FIRST_NAME_LENGTH_INDEX + _FIRST_NAME_LENGTH_SIZE


class BetPacketHeader:
    def __init__(self, first_name_length: int, last_name_length: int):
        self.first_name_length = first_name_length
        self.last_name_length = last_name_length

    @classmethod
    def from_bet(cls, bet: Bet) -> 'BetPacketHeader':
        first_name_length = len(bet.first_name.encode())
        last_name_length = len(bet.last_name.encode())

        if first_name_length > MAX_NAME_LENGTH:
            raise ValueError('first name is too long')

        if last_name_length > MAX_NAME_LENGTH:
            raise ValueError('last name is too long')

        return cls(first_name_length, last_name_length)

    def write_to_bytes(self, bytes: bytearray) -> None:
        bytes.append(self.first_name_length)
        bytes.append(self.last_name_length)

    @classmethod
    def from_bytes(cls, bytes: bytes) -> 'BetPacketHeader':
        if len(bytes) != BET_PACKET_HEADER_SIZE:
            raise ValueError('incorrect bet packet header length')

        first_name_length = int.from_bytes(
            bytes[_FIRST_NAME_LENGTH_INDEX: _FIRST_NAME_LENGTH_INDEX +
                  _FIRST_NAME_LENGTH_SIZE],
            byteorder=BYTE_ORDER,
        )
        last_name_length = int.from_bytes(
            bytes[_LAST_NAME_LENGTH_INDEX: _LAST_NAME_LENGTH_INDEX +
                  _LAST_NAME_LENGTH_SIZE],
            byteorder=BYTE_ORDER,
        )
        return cls(first_name_length, last_name_length)
