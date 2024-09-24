package algo

import (
	"strconv"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/message"
)

type VmData struct {
	Token grammar.Token
	Value int
}

func Parse(Token []Token) (data []VmData, err error) {
	for _, token := range Token {
		if token.Type == message.Error {
			err = token.Value.(*message.ErrorValue).Err
			return
		}
		switch token.Token {
		case grammar.Number:
			value, atoiErr := strconv.Atoi(string(token.Value.([]byte)))
			if atoiErr != nil {
				err = atoiErr
				return
			}
			data = append(data, VmData{
				Token: token.Token,
				Value: value,
			})
			continue
		default:
			data = append(data, VmData{
				Token: token.Token,
			})
		}
	}
	return
}
