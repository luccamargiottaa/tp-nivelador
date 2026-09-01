import socket
import logger
import safe_socket
from lottery import Bet, Lottery
from lottery.protocol import send_winners, recv_bets


class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str):
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(storage_path)

    def _handle_client(self, client_socket) -> None:
        action = 'recv-bets'

        try:
            logger.info(action, logger.LogResult.in_progress)

            while True:
                bets, agency_id = recv_bets(client_socket)

                if not bets:
                    break

                self.lottery.store_bets(bets)

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

            return

        action = 'send-winners'

        try:
            logger.info(action, logger.LogResult.in_progress)

            winners = self.get_winners(agency_id)
            send_winners(client_socket, agency_id, winners)

            logger.info(action, logger.LogResult.success)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, 'err', e)

    def get_winners(self, agency_id: int) -> list[Bet]:
        bets = self.lottery.load_bets()

        winners = list(filter(lambda bet: bet.agency_id ==
                              agency_id and self.lottery.has_won(bet), bets))

        return winners

    def run(self) -> None:
        action = "accept-connection"

        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()

            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()

                except Exception as e:
                    logger.error(action, logger.LogResult.fail)

                    raise e

                logger.info(action, logger.LogResult.success)
                self._handle_client(client_socket)
