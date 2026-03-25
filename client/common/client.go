package common

import (
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"encoding/csv"

	"github.com/op/go-logging"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/codec"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

var log = logging.MustGetLogger("log")
const MAX_BATCH_BYTES = 8 * 1024

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

	var betsPath = f"./data/agency-{c.config.ID}.csv"
	betsFile, err := os.Open(betsPath)
	if err != nil {
		log.Errorf("action: open_bets_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	reader := csv.NewReader(betsFile)
	bets, err := NextBatch(reader, c.config.ID)

	defer betsFile.Close()

	if err := protocol.SendMessage(c.conn, bets); err != nil  {
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

func NextBatch(reader *csv.Reader, clientID string) ([]model.ClientBet, error) {
	for i := 0; i < MAX_BATCH_BYTES; i++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading file: %w", err)
		}
		
		bet, err := parseBet(record, c.config.ID)
		if err != nil {
			log.Errorf("action: parse_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			continue
		}

		encodedBet := codec.EncodeBet(bet)
		// chequar si es necesario dividir en batches por ser muy grande?

		if currentBatchBytes+len(encoded) > MAX_BATCH_BYTES - protocol.HEADER_SIZE {
            break
        }

        bets = append(bets, bet)
        currentBatchBytes += len(encoded)
	}
	
	return bets, nil
}

func parseBet(record []string, clientID string) (model.ClientBet, error) {
    // Format: FirstName LastName,LastName,ID,BirthDate,Number
    
    number, err := strconv.Atoi(record[4])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid number: %w", err)
    }

    id, err := strconv.Atoi(record[2])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid ID: %w", err)
    }

    birthdate, err := time.Parse("2006-01-02", record[3])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid birthdate: %w", err)
    }

    return model.ClientBet{
        Agency:    clientID,
        Number:    number,
        Name:      record[0],
        Lastname:  record[1],
        ID:        id,        
        Birthdate: birthdate,
    }, nil
}