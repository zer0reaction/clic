// A really scuffed type system.

package types

type Id int

const IdNone Id = -1

type TypeNode struct {
	Tag   tag
	Size  uint
	Align uint

	// Type defenition
	DefinedAs Id

	// Struct & union
	Fields []Field
}

type Field struct {
	Type Id
	Name string // Empty string means anonymous
}

type tag uint

const (
	typeTagError tag = iota

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
var builtin = map[tag]Id{}

func init() {
	registerBuiltin(Void, 0)
	registerBuiltin(S64, 8)
	registerBuiltin(U64, 8)
	registerBuiltin(Bool, 1)
}

func Register(node TypeNode) Id {
	id := Id(len(table))
	table = append(table, node)
	return id
}

func registerBuiltin(tag tag, size uint) Id {
	_, ok := builtin[tag]
	if ok {
		panic("tried to register builtin type twice")
	}

	id := Id(len(table))
	table = append(table, TypeNode{
		Tag:   tag,
		Size:  size,
		Align: size,
	})
	builtin[tag] = id
	return id
}

func Get(id Id) TypeNode {
	return table[id]
}

func GetBuiltin(tag tag) Id {
	id, ok := builtin[tag]
	if !ok {
		panic("builtin type not registered")
		// id = Register(TypeNode{Tag: tag})
		// builtin[tag] = id
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
