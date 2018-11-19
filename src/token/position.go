package token

type Position struct {
	Line   int
	Column int
}

func (p *Position) AddLine() {
	p.Column = 0
	p.Line++
}

func (p *Position) AddColumn() {
	p.Column++
}
