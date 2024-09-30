package csv

type Option func(*Parser)

func WithHeader(header bool) Option {
	return func(p *Parser) {
		p.withHeader = header
	}
}

func WithSeparator(delimiter byte) Option {
	return func(p *Parser) {
		p.separator = delimiter
	}
}
