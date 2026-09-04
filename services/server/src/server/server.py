import socket
import logger
import safe_socket
import fcntl
import signal
from lottery import Bet, Lottery
from lottery.protocol import send_winners, recv_bets
from multiprocessing import Queue, Process
from collections import defaultdict
from typing import List, Iterator

TOKEN = 0


class WorkerState:
    def __init__(self):
        self.shutdown_done = False


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
        self._shutdown_done = False

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
        action = "close_request_queue"

        if self._request_queue is not None:
            try:
                self._request_queue.close()
                self._request_queue.join_thread()

                logger.info(action, logger.LogResult.success)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        action = "close_notification_queue"

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
                self._request_queue.put(None)

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
                worker.terminate()

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        for _ in range(len(self._workers)):
            try:
                self._notification_queue.put(None)

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        for worker in self._workers:
            try:
                worker.join()

            except Exception as e:
                logger.error(action, logger.LogResult.fail, 'err', e)

        logger.info(action, logger.LogResult.success)

    def _shutdown(self):
        if not self._shutdown_done:
            self._close_socket()
            self._join_coordinator()
            self._join_workers()
            self._close_queues()

            self._shutdown_done = True

    def _coordinator_shutdown(self):
        self._close_queues()

    @classmethod
    def _worker_shutdown(cls, client_socket: socket.socket, request_queue: Queue, notification_queue: Queue, state: WorkerState) -> None:
        if state.shutdown_done:
            return

        action = "worker-close-socket"

        try:
            client_socket.close()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

        action = "worker-close-request-queue"

        try:
            request_queue.close()
            request_queue.join_thread()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

        action = "worker-close-notification-queue"

        try:
            notification_queue.close()
            notification_queue.join_thread()

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

        state.shutdown_done = True

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

                    if agency_id is None:
                        self._coordinator_shutdown()

                        return

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

        if notification_queue.get() is None:
            return None

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

            if winners is None:
                return

            send_winners(client_socket, agency_id, winners)

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    @classmethod
    def _recv_bets(cls, client_socket: socket.socket, lot: Lottery, state: WorkerState) -> int:
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
            if not state.shutdown_done:
                logger.error(action, logger.LogResult.fail, 'err', e)

            raise e

    @classmethod
    def _handle_client(cls, client_socket: socket.socket, lot: Lottery, request_queue: Queue, notification_queue: Queue) -> None:
        logger.init()

        state = WorkerState()
        signal.signal(signal.SIGTERM, lambda signum, frame: cls._worker_shutdown(
            client_socket, request_queue, notification_queue, state))

        try:
            agency_id = cls._recv_bets(client_socket, lot, state)

            cls._send_winners(client_socket, agency_id,
                              request_queue, notification_queue, lot)

        except Exception as e:
            if not state.shutdown_done:
                raise e

        finally:
            cls._worker_shutdown(
                client_socket, request_queue, notification_queue, state)

    def _accept_loop(self):
        action = "accept-connection"

        while True:
            try:
                client_socket, _ = self._listen_socket.accept()

                logger.info(action, logger.LogResult.success)

                self._start_worker(client_socket)

            except Exception as e:
                if not self._shutdown_done:
                    logger.error(action, logger.LogResult.fail)

                raise e

    def run(self) -> None:
        signal.signal(signal.SIGTERM, lambda signum, frame: self._shutdown())

        try:
            self._open_queues()
            self._start_coordinator()
            self._open_socket()

            self._accept_loop()

        except Exception as e:
            if not self._shutdown_done:
                raise e

        finally:
            self._shutdown()
