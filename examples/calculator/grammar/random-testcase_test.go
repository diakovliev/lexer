package grammar

import (
	"fmt"
	"io"
)

type RandomTestCase struct {
	name    string
	content string
	input   io.Reader
	tokens  int
}

func NewRandomTestCase(opsCount uint, randomSpaces bool, randomScopes bool) (ret *RandomTestCase) {
	reader, size, tokens := GenerateRandomInput(opsCount, randomSpaces, randomScopes)
	ret = &RandomTestCase{
		name: fmt.Sprintf(
			"%d ops spaces: %t scopes %t size: %d tokens: %d",
			opsCount, randomSpaces, randomScopes, size, tokens,
		),
		input:   reader,
		tokens:  tokens,
		content: reader.String(),
	}
	return ret
}

func (rtc RandomTestCase) Name() string {
	return rtc.name
}

func (rtc RandomTestCase) Input() io.Reader {
	return rtc.input
}

func (rtc RandomTestCase) Tokens() int {
	return rtc.tokens
}

func (rtc RandomTestCase) Content() string {
	return rtc.content
}

func (rtc RandomTestCase) Size() int {
	return len(rtc.content)
}
