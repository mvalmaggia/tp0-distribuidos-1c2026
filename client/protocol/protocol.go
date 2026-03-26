package protocol

import (
	"io"
	"net"
	"fmt"
)

const HEADER_LENGTH = 8

// writeAll writes all data to the connection. It ensures that all bytes are sent, even if multiple Write calls are needed.
func writeAll(conn net.Conn, data []byte) error {
	totalSent := 0
    for totalSent < len(data) {
        n, err := conn.Write(data[totalSent:])
        if err != nil {
            return err
        }
        totalSent += n
    }
    return nil
}

// SendMessage sends a length-prefixed message over the connection.
func SendMessage(conn net.Conn, message string) error {
	data := []byte(message)
    header := []byte(fmt.Sprintf("%08d", len(data))) // 8-byte ASCII length
    fullMessage := append(header, data...)

    return writeAll(conn, fullMessage)
}

// ReceiveMessage reads a length-prefixed message from the connection and returns it as a string.
func ReceiveMessage(conn net.Conn) (string, error)	{
	header := make([]byte, HEADER_LENGTH)
    if _, err := io.ReadFull(conn, header); err != nil {
        return "", err
    }

	messageLength := 0
    if _, err := fmt.Sscanf(string(header), "%08d", &messageLength); err != nil {
        return "", fmt.Errorf("failed to parse length header: %w", err)
    }

	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(conn, messageData); err != nil {
		return "", err
	}

	return string(messageData), nil
}