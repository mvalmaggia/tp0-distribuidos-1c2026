package codec

import (
	"fmt"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

// EncodeBet Encodes a Bet struct into a string. The encoding format is:
// "agency|name|surname|document|birthDate|number". This function is used to encode
// the bet information before sending it to the server.
func EncodeBet(bet model.ClientBet) string {
	date := bet.BirthDate.Format("2006-01-02")

	return fmt.Sprintf("%s|%s|%s|%d|%s|%d", bet.Agency, bet.Name, bet.Surname, bet.Document, date, bet.Number)
}	

func EncodeBetBatch(bets []model.ClientBet) string {
    encodedBets := make([]string, len(bets))
    
    for i, bet := range bets {
        encodedBets[i] = EncodeBet(bet)
    }

    return "BET_BATCH\n" + strings.Join(encodedBets, "\n")
}

func DecodeWinners(encoded string) ([]int, error) {
    if strings.TrimSpace(encoded) == "" {
        return []int{}, nil
    }

    parts := strings.Split(encoded, "|")
    dnis := make([]int, len(parts))

    for i, part := range parts {
        fmt.Sscanf(strings.TrimSpace(part), "%d", &dnis[i])
    }
    return dnis, nil
}