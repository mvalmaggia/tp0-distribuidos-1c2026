HEADER_LENGTH = 8

def send_message(socket, message):
    msg_bytes = message.encode('utf-8')
    length = len(msg_bytes)

    header = length.to_bytes(HEADER_LENGTH, byteorder='big')
    socket.sendall(header + msg_bytes)

def receive_message(socket):
    header = socket.recv_all(HEADER_LENGTH)
    if not header:
        return None

    length = int.from_bytes(header, byteorder='big')
    msg_bytes = socket.recv_all(length)
    return msg_bytes.decode('utf-8')