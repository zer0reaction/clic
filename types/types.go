// A really scuffed type system.

package types

type Id int

const IdNone Id = -1

type TypeNode struct {
	Tag TypeTag

	// Type defenition
	DefinedAs Id

	// Struct & union
	Fields []Field
}

type Field struct {
	Type Id
	Name string // Empty string means anonymous
}

type TypeTag uint

const (
	typeTagError TypeTag = iota

	// Base types (defined here for simplicity)
	Void
	S64
	U64
	Bool

	// New type defenition
	Definition

	// Compound types
	Struct
)

var table = []TypeNode{}
var builtin = map[TypeTag]Id{}

func Register(node TypeNode) Id {
	id := Id(len(table))
	table = append(table, node)
	return id
}

func Get(id Id) TypeNode {
	return table[id]
}

func GetBuiltin(tag TypeTag) Id {
	id, ok := builtin[tag]
	if !ok {
		id = Register(TypeNode{Tag: tag})
		builtin[tag] = id
	}
	return id
}

func (id Id) Stringify() string {
	// Should not break on builtin types. We register the type
	// when we get an id.
	node := table[id]

	switch node.Tag {
	case Void:
		return "void"

	case S64:
		return "s64"

	case U64:
		return "u64"

	case Bool:
		return "bool"

	case Struct:
		s := "struct ("
		for _, field := range node.Fields {
			s += field.Name + ": " + field.Type.Stringify() + " "
		}
		s += ")"
		return s

	case Definition:
		return "type, defined as " + node.DefinedAs.Stringify()

	default:
		panic("not implemented")
	}
}
