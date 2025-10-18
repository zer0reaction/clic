// This file contains a symbol table implementation.

package symbol

import (
	"lisp-go/types"
)

type SymbolId uint

const (
	SymbolIdNone SymbolId = 0
)

type SymbolTag uint

const (
	symbolError SymbolTag = iota
	SymbolVariable
	SymbolFunction
	SymbolType
)

type Variable struct {
	Type   types.TypeId
	Offset uint // subtracted from RBP
}

type Function struct {
	Params []TypedIdent
}

type Type struct {
	Id types.TypeId
}

// Has 'Name' because it is not stored in the symbol table
type TypedIdent struct {
	Name string
	Type types.TypeId
}

type symbol struct {
	tag      SymbolTag
	name     string
	variable Variable
	function Function
}

type table struct {
	prev *table
	data map[string]SymbolId
}

var current = &table{
	prev: nil,
	data: make(map[string]SymbolId),
}
var storage = make(map[SymbolId]symbol)

func PushBlock() {
	current = &table{
		prev: current,
		data: make(map[string]SymbolId),
	}
}

func PopBlock() {
	if current.prev == nil {
		panic("popping global block table")
	}
	current = current.prev
}

func AddSymbol(name string, tag SymbolTag) SymbolId {
	_, ok := current.data[name]
	if ok {
		panic("symbol already exists in the current block")
	}

	s := symbol{
		name: name,
		tag:  tag,
	}
	id := SymbolId(len(storage) + 1)

	storage[id] = s
	current.data[name] = id

	return id
}

func ExistsAnywhere(name string, tag SymbolTag) bool {
	ptr := current

	for ptr != nil {
		_, ok := ptr.data[name]
		if !ok {
			ptr = ptr.prev
			continue
		} else {
			return true
		}
	}

	return false
}

func ExistsInBlock(name string, tag SymbolTag) bool {
	_, ok := current.data[name]
	return ok
}

func LookupAnywhere(name string, tag SymbolTag) SymbolId {
	ptr := current

	for ptr != nil {
		id, ok := ptr.data[name]
		if !ok {
			ptr = ptr.prev
			continue
		}

		s, ok := storage[id]
		if !ok {
			panic("symbol not found")
		}
		// TODO: Is this the right thing to do?
		if s.tag != tag {
			return SymbolIdNone
		}
		return id
	}
	return SymbolIdNone
}

func LookupInBlock(name string, tag SymbolTag) SymbolId {
	id, ok := current.data[name]
	if !ok {
		return SymbolIdNone
	}

	s, ok := storage[id]
	if !ok {
		panic("symbol not found")
	}
	// TODO: Is this the right thing to do?
	if s.tag != tag {
		return SymbolIdNone
	}
	return id
}

func GetName(id SymbolId) string {
	s, ok := storage[id]
	if !ok {
		panic("symbol doesn't exist")
	}
	if s.name == "" {
		panic("empty name")
	}
	return s.name
}

func SetVariable(id SymbolId, v Variable) {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolVariable {
		panic("symbol's type is not a variable")
	}

	s.variable = v
	storage[id] = s
}

func GetVariable(id SymbolId) Variable {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolVariable {
		panic("symbol's type is not a variable")
	}

	return s.variable
}

func SetFunction(id SymbolId, f Function) {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolFunction {
		panic("symbol's type is not a function")
	}

	s.function = f
	storage[id] = s
}

func GetFunction(id SymbolId) Function {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}
	if s.tag != SymbolFunction {
		panic("symbol's type is not a function")
	}

	return s.function
}
