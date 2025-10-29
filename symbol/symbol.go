package symbol

import (
	"clic/types"
)

type Id uint

const (
	IdNone Id = 0
)

type tag uint

const (
	symbolError tag = iota
	LocVar
	Function
	Type
)

type sLocVar struct {
	Type   types.Id
	Offset uint // subtracted from RBP
}

type sFunction struct {
	Params []TypedIdent
}

type sType struct {
	Id types.Id
}

// Has 'Name' because it is not stored in the symbol table
type TypedIdent struct {
	Name string
	Type types.Id
}

type symbol struct {
	Tag  tag
	Name string

	LocVar   sLocVar
	Function sFunction
	Type     sType
}

type table struct {
	prev *table
	data map[string]Id
}

var current = &table{
	prev: nil,
	data: make(map[string]Id),
}
var storage = make(map[Id]symbol)

func IsScopeGlobal() bool {
	return (current.prev == nil)
}

func PushBlock() {
	current = &table{
		prev: current,
		data: make(map[string]Id),
	}
}

func PopBlock() {
	if current.prev == nil {
		panic("popping global block table")
	}
	current = current.prev
}

func AddToBlock(name string, tag tag) Id {
	_, ok := current.data[name]
	if ok {
		panic("symbol already exists in the current block")
	}

	s := symbol{
		Name: name,
		Tag:  tag,
	}
	id := Id(len(storage) + 1)

	storage[id] = s
	current.data[name] = id

	return id
}

func Get(id Id) symbol {
	s, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}

	return s
}

func Set(id Id, new symbol) {
	_, ok := storage[id]

	if !ok {
		panic("symbol doesn't exist")
	}

	storage[id] = new
}

func ExistsAnywhere(name string, tag tag) bool {
	ptr := current

	for ptr != nil {
		id, ok := ptr.data[name]
		if !ok {
			ptr = ptr.prev
			continue
		}

		return (storage[id].Tag == tag)
	}

	return false
}

func ExistsInBlock(name string, tag tag) bool {
	id, ok := current.data[name]

	if !ok {
		return false
	}

	return (storage[id].Tag == tag)
}

func LookupAnywhere(name string, tag tag) Id {
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
			return IdNone
		}
		return id
	}

	return IdNone
}

func LookupInBlock(name string, tag tag) Id {
	id, ok := current.data[name]
	if !ok {
		return IdNone
	}

	s, ok := storage[id]
	if !ok {
		panic("symbol not found")
	}
	// TODO: Is this the right thing to do?
	if s.Tag != tag {
		return IdNone
	}
	return id
}
