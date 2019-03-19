package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
	Scope          SymbolScope

	outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol), Scope: GlobalScope}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	table := NewSymbolTable()
	table.outer = outer
	table.Scope = LocalScope
	return table
}

func (t *SymbolTable) Define(name string) Symbol {
	s := Symbol{Name: name, Index: t.numDefinitions, Scope: t.Scope}

	t.store[name] = s
	t.numDefinitions++
	return s
}

func (t *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	s := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	t.store[name] = s
	return s
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	if !ok && t.outer != nil {
		s, ok = t.outer.Resolve(name)
	}
	return s, ok
}
