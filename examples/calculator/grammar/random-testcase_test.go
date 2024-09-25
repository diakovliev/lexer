package grammar

import (
	"fmt"
	"io"
)

type RandomTestCase struct {
	name     string
	content  string
	input    io.Reader
	tokens   int
	maxDepth uint
}

func NewRandomTestCase(opsCount uint, randomSpaces bool, randomScopes bool, maxDepth uint) (ret *RandomTestCase) {
	reader, size, tokens := GenerateRandomInput(opsCount, randomSpaces, randomScopes, maxDepth)
	ret = &RandomTestCase{
		name: fmt.Sprintf(
			"%d ops spaces: %t scopes %t depth: %d size: %d tokens: %d",
			opsCount, randomSpaces, randomScopes, maxDepth, size, tokens,
		),
		input:    reader,
		tokens:   tokens,
		content:  reader.String(),
		maxDepth: maxDepth,
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

func (rtc RandomTestCase) MaxDepth() uint {
	return rtc.maxDepth
}
