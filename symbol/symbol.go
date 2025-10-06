// This file contains a symbol table implementation.

package symbol

type SymbolId uint

const (
	SymbolIdNone SymbolId = 0
)

type SymbolTag uint

const (
	symbolError SymbolTag = iota
	SymbolVariable
	SymbolFunction
)

type ValueType uint

const (
	valueError ValueType = iota
	ValueNone
	ValueS64
	ValueU64
	ValueBoolean
)

type Variable struct {
	Name   string
	Type   ValueType
	Offset uint // subtracted from RBP
}

type Function struct {
	Name string
}

type symbol struct {
	tag      SymbolTag
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

	s := symbol{tag: tag}
	id := SymbolId(len(storage) + 1)

	storage[id] = s
	current.data[name] = id

	return id
}

func LookupGlobal(name string, tag SymbolTag) SymbolId {
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
		// TODO: is this the right thing to do?
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
	// TODO: is this the right thing to do?
	if s.tag != tag {
		return SymbolIdNone
	}
	return id
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
