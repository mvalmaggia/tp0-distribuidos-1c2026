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

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)	

	go func() {
		sig := <-sigChan
		log.Infof("action: signal_received | result: in_progress | signal: %v", sig)
		c.running = false
		if c.conn != nil {
			c.conn.Close()
			log.Infof("action: shutdown_client_socket | result: success")
		}
		os.Exit(0)
	}()

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
			break
		}

		encodedBetsBatch := codec.EncodeBetBatch(betsBatch)

		for attempt := 1; attempt <= MAX_RETRY_ATTEMPTS; attempt++ {
		
			response, err := SendBetBatch(c, encodedBetsBatch)
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
					break
			} else if strings.TrimSpace(response) == "ERROR" {
				log.Errorf("action: apuesta_batch_enviada | result: fail | client_id: %v | batch_size: %v",
					c.config.ID, len(betsBatch))
					time.Sleep(2 * time.Second) 
			}
		}
	}
}

func SendBetBatch(client *Client, encodedBetsBatch string) (string, error) {
	if err := client.createClientSocket(); err != nil {
		return "", err
	}

	if err := protocol.SendMessage(client.conn, encodedBetsBatch); err != nil  {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
				client.config.ID,
				err,
			)
			client.conn.Close()
			return "", err
	}

	msg, err := protocol.ReceiveMessage(client.conn)
	client.conn.Close()

	return msg, err
}