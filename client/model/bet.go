package model

type Bet struct {
	Name      string
	Surname   string
	Document  string
	BirthDate string
	Number    string
}

func NewBet(name string, surname string, document string, birthDate string, number string) *Bet {
	return &Bet{
		Name:      name,
		Surname:   surname,
		Document:  document,
		BirthDate: birthDate,
		Number:    number,
	}
}
