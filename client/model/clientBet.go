package model

import (
	"time"
	"os"
	
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientBet struct {
	Name      string
	Surname   string
	Document  string
	BirthDate time.Time
	Number    string
}

func NewClientBet(name string, surname string, document string, birthDateStr string, number string) *ClientBet {
	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		log.Criticalf("Could not parse NACIMIENTO as date: %v", err)
		os.Exit(1)
	}

	return &ClientBet{
		Name:      name,
		Surname:   surname,
		Document:  document,
		BirthDate: birthDate,
		Number:    number,
	}
}
