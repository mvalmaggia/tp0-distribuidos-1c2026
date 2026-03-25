import socket
import logging
import signal
import sys
import protocol.protocol as protocol
from codec.codec import decode_bet
from common.utils import store_bets

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

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

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = protocol.receive_message(client_sock)
            client_bet = decode_bet(msg)

            store_bets([client_bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {client_bet.document} | numero: {client_bet.number}')
            
            addr = client_sock.getpeername()
            protocol.send_ack(client_sock)
            logging.info(f'action: send_ack | result: success | ip: {addr[0]}')

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try: 
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        except socket.timeout:
            return None
        
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
