package common

import (
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strings"

	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

var log = logging.MustGetLogger("log")

const MAX_RETRY_ATTEMPTS = 3
const MAX_AMOUNT_POLLS = 5

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		running: true,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func setupSignalHandler(c *Client) {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

    go func() {
        sig := <-sigs
        log.Infof("action: signal_received | result: in_progress | signal: %v", sig)
        c.running = false
        if c.conn != nil {
            c.conn.Close()
            log.Infof("action: shutdown_client_socket | result: success")
        }
        os.Exit(0)
    }()
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClient() {
	setupSignalHandler(c)

	betsParser, err := NewBetParser(c.config.ID, "./dataset.csv")
	if err != nil {
		log.Errorf("action: open_bets_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	defer betsParser.Close()

	for c.running {
		betsBatch, err := betsParser.NextBatch(c.config.BatchMaxAmount)
		if err != nil {
			log.Errorf("action: read_bets_batch | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			continue
		}

		if len(betsBatch) == 0 {
			log.Infof("action: all_bets_sent | result: success | client_id: %v", c.config.ID)
			HandleEndOfBatch(c)
			break
		}

		encodedBetsBatch := codec.EncodeBetBatch(betsBatch)

		for attempt := 1; attempt <= MAX_RETRY_ATTEMPTS; attempt++ {
		
			response, err := SendMessageToConnection(c, encodedBetsBatch)
			if err != nil {
				log.Errorf("action: apuesta_batch_enviada | result: fail | client_id: %v | attempt: %v | error: %v",
					c.config.ID,
					attempt,
					err,
				)
				time.Sleep(2 * time.Second)
				continue
			}
			
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)

			if err != nil {
				log.Errorf("action: apuesta_batch_enviada | result: fail | batch_size: %v",
					len(betsBatch),
				)
				continue
			}

			if strings.TrimSpace(response) == "ACK" {
				log.Infof("action: apuesta_batch_enviada | result: success | client_id: %v | batch_size: %v",
					c.config.ID, len(betsBatch))
			} else if strings.TrimSpace(response) == "ERROR" {
				log.Errorf("action: apuesta_batch_enviada | result: fail | client_id: %v | batch_size: %v",
					c.config.ID, len(betsBatch))
			}
		}
	}
}

func HandleEndOfBatch(c *Client) {

	if err := SendMessageToConnection(c, fmt.Sprintf("BATCH_END:%s", c.config.ID)); err != nil {
		log.Errorf("action: send_end_of_batch | result: fail | error: %v", err)
		return
	}

    polls := 0
    for polls < MAX_AMOUNT_POLLS {
        if err := c.createClientSocket(); err != nil {
            log.Errorf("action: handle_end_of_batch | result: fail | error: %v", err)
            return
        }

        if err := protocol.SendMessage(c.conn, fmt.Sprintf("GET_WINNERS:%s", c.config.ID)); err != nil {
            log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
            c.conn.Close()
            return
        }

		if err := SendMessageToConnection(c, fmt.Sprintf("GET_WINNERS:%s", c.config.ID)); err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
			c.conn.Close()
			return
		}
    }
	log.Infof("action: consulta_ganadores | result: fail | reason: max_polls_reached")
} 

func SendMessageToConnection(client *Client, message string) (string, error) {
	if err := client.createClientSocket(); err != nil {
		return "", err
	}

	if err := protocol.SendMessage(client.conn, message); err != nil  {
		return "", err
	}

	response, err := protocol.ReceiveMessage(client.conn)
	client.conn.Close()

	return response, err
}