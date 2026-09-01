from .constants import BYTE_ORDER

_CODE_SIZE = 1
_BET_AMOUNT_SIZE = 1
_BET_SIZE_SIZE = 2
_AGENCY_ID_SIZE = 1
HEADER_SIZE = _CODE_SIZE + _BET_AMOUNT_SIZE + _BET_SIZE_SIZE + _AGENCY_ID_SIZE

_CODE_INDEX = 0
_BET_AMOUNT_INDEX = _CODE_INDEX + _CODE_SIZE
_BET_SIZE_INDEX = _BET_AMOUNT_INDEX + _BET_AMOUNT_SIZE
_AGENCY_ID_INDEX = _BET_SIZE_INDEX + _BET_SIZE_SIZE

ACK_CODE = 0
BET_CODE = 1
END_CODE = 2


class Header:
    def __init__(
            self,
            code: int,
            bet_amount: int,
            bet_size: int,
            agency_id: int):
        self.code = code
        self.bet_amount = bet_amount
        self.bet_size = bet_size
        self.agency_id = agency_id

    @classmethod
    def _new_non_bets_header(cls, code: int, agency_id: int) -> 'Header':
        return cls(code, 0, 0, agency_id)

    @classmethod
    def new_ack_header(cls, agency_id: int) -> 'Header':
        return cls._new_non_bets_header(ACK_CODE, agency_id)

    @classmethod
    def new_bet_header(cls, bet_amount: int, bet_size: int, agency_id: int) -> 'Header':
        return cls(BET_CODE, bet_amount, bet_size, agency_id)

    @classmethod
    def new_end_header(cls, agency_id: int) -> 'Header':
        return cls._new_non_bets_header(END_CODE, agency_id)

    def write_to_bytes(self, bytes: bytearray) -> None:
        bytes.append(self.code)
        bytes.append(self.bet_amount)

        bytes.extend(
            self.bet_size.to_bytes(
                length=_BET_SIZE_SIZE,
                byteorder=BYTE_ORDER))

        bytes.append(self.agency_id)

    @classmethod
    def from_bytes(cls, bytes: bytes) -> 'Header':
        if len(bytes) != HEADER_SIZE:
            raise ValueError('invalid header size')

        code = bytes[_CODE_INDEX]
        bet_amount = bytes[_BET_AMOUNT_INDEX]

        bet_size = int.from_bytes(
            bytes[_BET_SIZE_INDEX: _BET_SIZE_INDEX +
                  _BET_SIZE_SIZE], byteorder=BYTE_ORDER
        )
        agency_id = bytes[_AGENCY_ID_INDEX]

        if code == BET_CODE:
            if bet_amount == 0:
                raise ValueError('invalid bet amount')

            if bet_size == 0:
                raise ValueError('invalid bet size')

        elif bet_amount != 0 or bet_size != 0:
            raise ValueError('invalid header code')

        return cls(code, bet_amount, bet_size, agency_id)
