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
	Variable
	Function
	Type
)

type sVariable struct {
	Type   types.TypeId
	Offset uint // subtracted from RBP
}

type sFunction struct {
	Params []TypedIdent
}

type sType struct {
	Id types.TypeId
}

// Has 'Name' because it is not stored in the symbol table
type TypedIdent struct {
	Name string
	Type types.TypeId
}

type symbol struct {
	Tag  SymbolTag
	Name string

	Variable sVariable
	Function sFunction
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

func AddToBlock(name string, tag SymbolTag) SymbolId {
	_, ok := current.data[name]
	if ok {
		panic("symbol already exists in the current block")
	}

	s := symbol{
		Name: name,
		Tag:  tag,
	}
	id := SymbolId(len(storage) + 1)

	storage[id] = s
	current.data[name] = id

	return id
}

func Get(id SymbolId) symbol {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}

	return s
}

func Set(id SymbolId, new symbol) {
	_, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}

	storage[id] = new
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
		if s.Tag != tag {
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
	if s.Tag != tag {
		return SymbolIdNone
	}
	return id
}
