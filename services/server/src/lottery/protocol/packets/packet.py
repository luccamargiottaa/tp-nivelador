from typing import List

from lottery import Bet
from lottery.protocol.packets.bet_packet.bet_packet import BetPacket
from lottery.protocol.packets.header import Header

MAX_BETS = 255


class Packet:
    def __init__(self, header: Header, bets: List[BetPacket] = []):
        self.header = header
        self.bets = bets

    @classmethod
    def new_ack_packet(cls, agency_id: int) -> 'Packet':
        return cls(Header.new_ack_header(agency_id))

    @classmethod
    def new_end_packet(cls, agency_id: int) -> 'Packet':
        return cls(Header.new_end_header(agency_id))

    @classmethod
    def new_bet_packet(cls, agency_id: int, bets: List[Bet]) -> 'Packet':
        bet_amount = len(bets)

        if bet_amount == 0:
            raise ValueError('bets is empty')

        if bet_amount > MAX_BETS:
            raise ValueError('too many bets')

        bet_packets = []
        bet_size = 0

        for bet in bets:
            bet_packet = BetPacket.from_bet(bet)
            bet_packets.append(bet_packet)
            bet_size += bet_packet.size()

        header = Header.new_bet_header(bet_amount, bet_size, agency_id)

        return cls(header, bet_packets)

    def to_bytes(self) -> bytearray:
        bytes = bytearray()

        self.header.write_to_bytes(bytes)

        for bet in self.bets:
            bet.write_to_bytes(bytes)

        return bytes

    @classmethod
    def add_bets_from_bytes(cls, header: Header, bytes: bytes, bets: List[Bet]) -> None:
        offset = 0

        while offset < header.bet_size:
            packet = BetPacket.from_bytes(bytes[offset:], header.agency_id)
            bets.append(packet.bet)
            offset += packet.size()
