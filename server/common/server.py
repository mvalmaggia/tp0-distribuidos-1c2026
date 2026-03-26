import socket
import logging
import signal
import time
import protocol.protocol as protocol
from codec.codec import decode_bet_batch, encode_winners
from common.utils import store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog, expected_agencies=5):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._finished_agencies = set()
        self._expected_agencies = expected_agencies
        self._winners_by_agency = {}

        self._server_socket.settimeout(1)

        self._running = True
        self._client_sock = None    

        signal.signal(signal.SIGINT, self.handle_signal)
        signal.signal(signal.SIGTERM, self.handle_signal)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communication
        finishes, servers starts to accept new connections again
        """

        while self._running:
            self._client_sock = self.__accept_new_connection()
            if self._client_sock:
                self.__handle_client_connection(self._client_sock)
                self._client_sock = None
        self.__graceful_shutdown()

    def _get_winners_by_agency(self):
        self._winners_by_agency = {}
        bets = load_bets()
        for bet in bets:
            if has_won(bet):
                agency = bet.agency
                if agency not in self._winners_by_agency:
                    self._winners_by_agency[agency] = []
                self._winners_by_agency[agency].append(bet.document)

    def _handle_batch_bet(self, encoded_msg):
        bets = decode_bet_batch(encoded_msg)

        store_bets(bets)
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

    def _get_winners_for_agency(self, agency_id: int):
        """
        Retrieves a list of document ids for all the winning bets
        for a specific agency.
        """
        logging.info(f"action: get_winners_for_agency | result: in_progress | agency: {agency_id}")
        return self._winners_by_agency.get(agency_id, [])

    def _handle_get_winners(self, client_sock, agency_id):
        if len(self._finished_agencies) != self._expected_agencies:
            protocol.send_message(client_sock, "ERROR:NOT_ALL_BATCHES_RECEIVED")
            return

        winners = self._get_winners_for_agency(agency_id)
        protocol.send_message(client_sock, encode_winners(winners))
        logging.info(f"action: send_winners | result: success | agency: {agency_id}")

    def _handle_end_of_batch(self, agency_id):
        self._finished_agencies.add(agency_id)
        logging.info(f"action: batch_end_received | result: success | agency: {agency_id}")

        if len(self._finished_agencies) >= self._expected_agencies:
            logging.info("action: sorteo | result: success")
            self._get_winners_by_agency()

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = protocol.receive_message(client_sock)
            if msg.startswith("BET_BATCH"):
                self._handle_batch_bet(msg)

            if msg.startswith("GET_WINNERS"):
                agency_id = msg.split(":", 1)[1].strip()
                self._handle_get_winners(client_sock, agency_id)
                return
            
            if msg.startswith("BATCH_END"):
                agency_id = msg.split(":", 1)[1].strip()
                self._handle_end_of_batch(agency_id)

            addr = client_sock.getpeername()
            protocol.send_ack(client_sock)
            logging.info(f'action: send_ack | result: success | ip: {addr[0]}')

        # except (ValueError, IndexError) as e:
        #     logging.error(f"action: apuesta_recibida | result: fail")
        #     try:
        #         protocol.send_message(client_sock, "ERROR")
        #     except OSError:
        #         pass
        except OSError as e:
            logging.error(f"action: apuesta_recibida | result: fail")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        logging.info('action: accept_connections | result: in_progress')
        # Connection arrived
        while self._running:
            try: 
                c, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                time.sleep(0.5)
                return c
            except socket.timeout:
                continue
        
        return c
    
    def __graceful_shutdown(self):
        
        if self._client_sock:
            self._client_sock.shutdown(socket.SHUT_RDWR)
            self._client_sock.close()
            logging.info('action: shutdown_client_socket | result: success')
    
        if self._server_socket:
            self._server_socket.close()
            logging.info("action: shutdown_server_socket | result: success")

    def handle_signal(self, signum=None, frame=None):
        logging.info(f'action: signal_received | result: success')
        self._running = False
