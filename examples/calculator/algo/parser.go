package algo

import (
	"strconv"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/message"
)

// Parse - parse tokens to VMCode
func Parse(tokens []Token) (data []VMCode, err error) {
	data = make([]VMCode, 0, len(tokens))
	for _, token := range tokens {
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
			data = append(data, VMCode{Token: token.Token, Value: value})
			continue
		default:
			data = append(data, VMCode{Token: token.Token})
		}
	}
	return
}
