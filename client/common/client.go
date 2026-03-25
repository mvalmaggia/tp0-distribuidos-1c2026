package common

import (
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strings"

	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
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
func (c *Client) StartClientLoop(clientBet *model.ClientBet) {
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

	c.createClientSocket()

	encodedBet := codec.EncodeBet(clientBet)
	if err := protocol.SendMessage(c.conn, encodedBet); err != nil  {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.conn.Close()
		return
	}

	response, err := protocol.ReceiveMessage(c.conn)
	c.conn.Close()
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		
		return
	}

	if strings.TrimSpace(response) == "ACK" {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
				clientBet.Document,
				clientBet.Number,
			)
	}	
}
