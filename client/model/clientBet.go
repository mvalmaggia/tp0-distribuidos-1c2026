package model

import (
	"time"
	
	"github.com/op/go-logging"
)

type ClientBet struct {
	Name      string
	Surname   string
	Document  string
	BirthDate string
	Number    time.Time
}

func NewClientBet(name string, surname string, document string, birthDate string, number string) *Bet {
	birthDate, err := time.Parse("2006-01-02", nacimientoStr)
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
