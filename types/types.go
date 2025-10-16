// A really scuffed type system. Type is represented as an index in
// the type table, so you can just compare indexes to see if the types
// match or not. Types are not represented as pointers, because this
// will introduce a lot of bugs.

package types

// TODO: Add ToString() and add to error reports

type Type uint

// Can be later registered as a type, but you need to register all the
// nested types recursively first
type TypeDummy struct {
	Tag TypeTag

	// Struct & union
	Fields []Type
}

// If you want to get the actual 'Type', use GetBuiltin()
type TypeTag uint

const (
	typeError TypeTag = iota
	None
	S64
	U64
	Bool
	Struct
	Union
)

var table = []TypeDummy{}
var builtins = make(map[TypeTag]Type)

func GetBuiltin(tag TypeTag) Type {
	id, ok := builtins[tag]
	if !ok {
		id = Type(len(table))
		table = append(table, TypeDummy{Tag: tag})
		builtins[tag] = id
	}
	return id
}

// Assumes dummy is a unique type. Sigh. Should recursively hash the
// type tags or something.
func Register(dummy TypeDummy) Type {
	id := Type(len(table))
	table = append(table, dummy)
	return id
}
