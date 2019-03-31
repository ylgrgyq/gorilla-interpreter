package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
	Function     SymbolScope = "FUNCTION"
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
	FreeSymbols    []Symbol

	outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol), Scope: GlobalScope, FreeSymbols: nil}
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

func (t *SymbolTable) DefineFunctionName(name string) Symbol {
	s := Symbol{Name: name, Index: 0, Scope: Function}

	t.store[name] = s
	return s
}

func (t *SymbolTable) defineFree(original Symbol) Symbol {
	t.FreeSymbols = append(t.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(t.FreeSymbols) - 1, Scope: FreeScope}

	t.store[original.Name] = symbol
	return symbol
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	if !ok && t.outer != nil {
		s, ok = t.outer.Resolve(name)
		if !ok {
			return s, ok
		}

		if s.Scope == GlobalScope || s.Scope == BuiltinScope {
			return s, ok
		}

		free := t.defineFree(s)
		return free, true
	}
	return s, ok
}
