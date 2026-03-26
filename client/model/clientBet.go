package model

import (
	"time"
	"os"
	
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientBet struct {
	Agency	  string
	Name      string
	Surname   string
	Document  int
	BirthDate time.Time
	Number    int
}

// NewClientBet creates a new ClientBet instance from the provided parameters.
func NewClientBet(agency string, name string, surname string, document int, birthDateStr string, number int) ClientBet {
	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		log.Criticalf("Could not parse NACIMIENTO as date: %v", err)
		os.Exit(1)
	}

	return ClientBet{
		Agency:	  agency,
		Name:      name,
		Surname:   surname,
		Document:  document,
		BirthDate: birthDate,
		Number:    number,
	}
}
