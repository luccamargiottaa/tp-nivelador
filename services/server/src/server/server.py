import socket
import logger
import safe_socket
import fcntl
from lottery import Bet, Lottery
from lottery.protocol import send_winners, recv_bets
from multiprocessing import Queue, Process
from collections import defaultdict
from typing import List, Iterator

TOKEN = 0


class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str, agency_quorum_min: int):
        self._server_host = server_host
        self._server_port = server_port
        self._lottery = Lottery(storage_path)
        self._agency_quorum_min = agency_quorum_min
        self._request_queue = None
        self._notification_queue = None
        self._coordinator = None
        self._listen_socket = None
        self._workers = []

    def _open_queues(self) -> None:
        action = "open_queues"

        try:
            self._request_queue = Queue()
            self._notification_queue = Queue()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    def _close_queues(self) -> None:
        action = "close_queues"

        if self._request_queue is not None:
            try:
                self._request_queue.close()
                self._request_queue.join_thread()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        if self._notification_queue is not None:
            try:
                self._notification_queue.close()
                self._notification_queue.join_thread()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

    def _start_coordinator(self) -> None:
        action = "start-coordinator"

        try:
            self._coordinator = Process(target=self._coordinator_loop)
            self._coordinator.start()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    def _join_coordinator(self) -> None:
        action = "join-coordinator"

        if self._coordinator is not None:
            try:
                self._coordinator.join()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

    def _open_socket(self) -> None:
        action = "open_socket"

        try:
            self._listen_socket = socket.socket(
                socket.AF_INET, socket.SOCK_STREAM)
            self._listen_socket.bind((self._server_host, self._server_port))
            self._listen_socket.listen()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    def _close_socket(self) -> None:
        action = "close_socket"

        if self._listen_socket is not None:
            try:
                self._listen_socket.close()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

    def _start_worker(self, client_socket: socket.socket) -> None:
        action = "start-worker"

        try:
            worker = Process(target=self._handle_client, args=(
                client_socket, self._lottery, self._request_queue, self._notification_queue))
            worker.start()
            self._workers.append(worker)

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    def _join_workers(self) -> None:
        action = "join-workers"

        for worker in self._workers:
            try:
                worker.join()

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        logger.info(action, logger.LogResult.success)

    @classmethod
    def _store_bets_locked(cls, bets: List[Bet], lot: Lottery) -> None:
        with open(lot.storage_path, "a+") as storage:
            fcntl.flock(storage, fcntl.LOCK_EX)

            try:
                lot.store_bets(bets)

            finally:
                fcntl.flock(storage, fcntl.LOCK_UN)

    @classmethod
    def _load_bets_locked(cls, lot: Lottery) -> Iterator[Bet]:
        with open(lot.storage_path, "r") as storage:
            fcntl.flock(storage, fcntl.LOCK_SH)

            try:
                yield from lot.load_bets()

            finally:
                fcntl.flock(storage, fcntl.LOCK_UN)

    def _coordinator_loop(self) -> None:
        logger.init()

        agencies = set()

        while True:
            action = "wait-for-quorum"

            logger.info(action, logger.LogResult.in_progress)

            while len(agencies) < self._agency_quorum_min:
                try:
                    agency_id = self._request_queue.get()

                except Exception as e:
                    logger.error(action, logger.LogResult.fail, 'err', e)

                    raise e

                agencies.add(agency_id)

            logger.info(action, logger.LogResult.success)

            notified_agencies = "-".join(str(agency) for agency in agencies)
            logger.info(action, logger.LogResult.in_progress,
                        "agencies", notified_agencies)

            for _ in range(self._agency_quorum_min):
                try:
                    self._notification_queue.put(TOKEN)

                except Exception as e:
                    logger.error(action, logger.LogResult.fail, 'err', e)

                    raise e

            logger.info(action, logger.LogResult.success,
                        "agencies", notified_agencies)

            agencies.clear()

    @classmethod
    def _get_winners(cls, agency_id: int, request_queue: Queue, notification_queue: Queue, lot: Lottery) -> list[Bet]:
        request_queue.put(agency_id)

        notification_queue.get()

        winners = []

        for bet in cls._load_bets_locked(lot):
            if lot.has_won(bet) and bet.agency_id == agency_id:
                winners.append(bet)

        return winners

    @classmethod
    def _send_winners(cls, client_socket: socket.socket, agency_id: int, request_queue: Queue, notification_queue: Queue, lot: Lottery) -> None:
        action = 'send-winners'

        try:
            logger.info(action, logger.LogResult.in_progress)

            winners = cls._get_winners(
                agency_id, request_queue, notification_queue, lot)
            send_winners(client_socket, agency_id, winners)

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    @classmethod
    def _recv_bets(cls, client_socket: socket.socket, lot: Lottery) -> int:
        action = 'recv-bets'

        try:
            logger.info(action, logger.LogResult.in_progress)

            while True:
                bets, agency_id = recv_bets(client_socket)

                if not bets:
                    break

                cls._store_bets_locked(bets, lot)

            logger.info(action, logger.LogResult.success)

            return agency_id

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    @classmethod
    def _handle_client(cls, client_socket: socket.socket, lot: Lottery, request_queue: Queue, notification_queue: Queue) -> None:
        logger.init()

        try:
            agency_id = cls._recv_bets(client_socket, lot)

            cls._send_winners(client_socket, agency_id,
                              request_queue, notification_queue, lot)

        finally:
            client_socket.close()

    def _accept_loop(self):
        action = "accept-connection"

        while True:
            try:
                client_socket, _ = self._listen_socket.accept()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail)

                raise e

            self._start_worker(client_socket)

    def run(self) -> None:
        try:
            self._open_queues()
            self._start_coordinator()
            self._open_socket()

            self._accept_loop()

        finally:
            self._close_socket()
            self._close_queues()
            self._join_coordinator()
            self._join_workers()
