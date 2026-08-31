import socket
from typing import List

import safe_socket
from lottery import Bet
from lottery.protocol.packets.header import (
    ACK_CODE,
    BET_CODE,
    END_CODE,
    HEADER_SIZE,
    Header,
)
from lottery.protocol.packets.packet import Packet

from services.server.src.lottery.protocol.packets.packet import MAX_BETS


def _send_packet(socket: socket.socket, packet: Packet) -> None:
    safe_socket.send_all(socket, packet.to_bytes())


def _recv_header(socket: socket.socket) -> Header:
    bytes = safe_socket.recv_all(socket, HEADER_SIZE)

    if not bytes:
        raise ConnectionError('connection closed while receiving header')

    header = Header.from_bytes(bytes)

    return header


def _expect_ack(socket: socket.socket, agency_id: int) -> None:
    header = _recv_header(socket)

    if header.code != ACK_CODE:
        raise ValueError('unexpected code from client')

    if header.agency_id != agency_id:
        raise ValueError('unexpected agency id')


def send_winners(socket: socket.socket, agency_id: int, winners: List[Bet]) -> None:
    offset = 0

    while offset < len(winners):
        bet_packet = Packet.new_bet_packet(agency_id, winners[offset:offset + MAX_BETS])
        _send_packet(socket, bet_packet)
        _expect_ack(socket, agency_id)

        offset += MAX_BETS

    end_packet = Packet.new_end_packet(agency_id)
    _send_packet(socket, end_packet)
    _expect_ack(socket, agency_id)


def recv_bets(socket: socket.socket) -> List[Bet]:
    header = _recv_header(socket)

    if header.code == END_CODE:
        _send_packet(socket, ack)
        return []

    if header.code != BET_CODE:
        raise ValueError('unexpected message from client')

    bytes = safe_socket.recv_all(socket, header.bet_size)

    if not bytes:
        raise ConnectionError('connection closed while receiving bets')

    bets = []
    Packet.add_bets_from_bytes(header, bytes, bets)

    _send_packet(socket, ack)

    return bets
