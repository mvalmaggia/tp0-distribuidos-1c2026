package codec

import (
	"fmt"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

// EncodeBet Encodes a Bet struct into a string. The encoding format is:
// "agency|name|surname|document|birthDate|number". This function is used to encode
// the bet information before sending it to the server.
func EncodeBet(bet *model.ClientBet) string {
	date := bet.BirthDate.Format("2006-01-02")

	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", bet.Agency, bet.Name, bet.Surname, bet.Document, date, bet.Number)
}	