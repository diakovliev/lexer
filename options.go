package lexer

type Characher interface {
	byte | rune
}

type Options[C Characher] struct {
	// symbols what are be treated as a spaces
	spaces []C
	// new line symbol
	nl C
}

type Option[C Characher] func(options *Options[C])

func WithSpaces[C Characher](spaces []C) Option[C] {
	return func(options *Options[C]) {
		options.spaces = spaces
	}
}

func WithNewLine[C Characher](nl C) Option[C] {
	return func(options *Options[C]) {
		options.nl = nl
	}
}
