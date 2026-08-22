package parser

import (
	"context"
	"time"
)

type Parser struct {
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, text string, now time.Time) (*Result, error) {

	return ParseDeterministic(text, now)
}
