package symbol

import (
	"errors"
)

type SymbolId int
type BlockId int

const (
	SymbolIdNone SymbolId = -1
	BlockIdNone  BlockId  = -1
)

type SymbolTag uint

const (
	SymbolVariable SymbolTag = iota
	SymbolFunction
)

type Variable struct {
	BlockId BlockId
	Name    string
	Offset  uint // subtracted from RBP
}

type Function struct {
	Name string
}

type symbol struct {
	tag      SymbolTag
	variable Variable
	function Function
}

var table = make(map[SymbolId]symbol)

func AddSymbol(tag SymbolTag) SymbolId {
	s := symbol{tag: tag}
	id := SymbolId(len(table))
	table[id] = s
	return id
}

func SetVariable(id SymbolId, v Variable) {
	s, ok := table[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolVariable {
		panic("symbol's type is not a variable")
	}

	s.variable = v
	table[id] = s
}

func GetVariable(id SymbolId) Variable {
	s, ok := table[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolVariable {
		panic("symbol's type is not a variable")
	}

	return s.variable
}

func LookupVariable(name string, blockId BlockId) (SymbolId, error) {
	for id, s := range table {
		nameMatch := (s.variable.Name == name)
		blockMatch := (s.variable.BlockId == blockId)
		if nameMatch && blockMatch {
			return id, nil
		}
	}
	return SymbolIdNone, errors.New("internal: variable not found")
}

func SetFunction(id SymbolId, f Function) {
	s, ok := table[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolFunction {
		panic("symbol's type is not a function")
	}

	s.function = f
	table[id] = s
}

func GetFunction(id SymbolId) Function {
	s, ok := table[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolFunction {
		panic("symbol's type is not a function")
	}

	return s.function
}

func LookupFunction(name string) (SymbolId, error) {
	for id, s := range table {
		nameMatch := (s.function.Name == name)
		if nameMatch {
			return id, nil
		}
	}
	return SymbolIdNone, errors.New("internal: function not found")
}
