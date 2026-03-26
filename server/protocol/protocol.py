HEADER_LENGTH = 8

def send_ack(socket):
    send_message(socket, "ACK")

def send_message(socket, message):
    data = message.encode("utf-8")
    length_str = f"{len(data):0{HEADER_LENGTH}}".encode("utf-8")
    socket.sendall(length_str + data)

def receive_message(socket) -> str:
    """
    Receive a message with length header.
    """
    # Read the header
    header = read_bytes(socket, HEADER_LENGTH)
    length = int(header.decode("utf-8"))

    # Read the payload
    payload = read_bytes(socket, length)
    msg = payload.decode("utf-8")
    return payload.decode("utf-8")

def read_bytes(sock, length):
    buffer = b""
    while len(buffer) < length:
        chunk = sock.recv(length - len(buffer))
        if not chunk:
            raise ConnectionError("Connection closed while reading payload")
        buffer += chunk

    return buffer