package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol)}
}

func (t *SymbolTable) Define(name string) Symbol {
	s := Symbol{Name: name, Index: t.numDefinitions, Scope: GlobalScope}

	t.store[name] = s
	t.numDefinitions++
	return s
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	return s, ok
}
