package protocol

import (
	"io"
	"net"
	"fmt"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

HEADER_LENGTH = 8

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

func SendMessage(conn net.Conn, message string) error {
	data := []byte(message)
    header := []byte(fmt.Sprintf("%08d", len(data))) // 8-byte ASCII length
    fullMessage := append(header, data...)

    return writeAll(conn, fullMessage)
}

func ReceiveMessage(conn net.Conn) (string, error)	{
	header := make([]byte, HEADER_LENGTH)
    if _, err := io.ReadFull(conn, header); err != nil {
        return nil, err
    }

	messageLength := binary.BigEndian.Uint32(header)
	log.Infof("Received header: %s, message length: %d", string(header), messageLength)	

	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(conn, messageData); err != nil {
		return nil, err
	}

	return string(messageData), nil
}