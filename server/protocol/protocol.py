HEADER_LENGTH = 8

def send_ack(socket):
    send_message(socket, "ACK")

def send_message(socket, message):
    msg_bytes = message.encode('utf-8')
    length = len(msg_bytes)

    header = length.to_bytes(HEADER_LENGTH, byteorder='big')
    socket.sendall(header + msg_bytes)

def recv_all(sock, n):
    data = bytearray()
    while len(data) < n:
        packet = sock.recv(n - len(data))
        if not packet: 
            return None 
        data.extend(packet)
    return data

def receive_message(socket):
    header = recv_all(socket, HEADER_LENGTH)
    if not header:
        return None

    length = int.from_bytes(header, byteorder='big')
    msg_bytes = recv_all(socket, length)
    return msg_bytes.decode('utf-8')

