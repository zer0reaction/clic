// A really scuffed type system. Type is represented as a hash, and
// stored in a map. Will break things if there is a hash collision.

package types

// TODO: Add ToString() and add to error reports

type TypeHash uint32

// Can be later registered as a type, but you need to register all the
// nested types recursively first
type TypeDummy struct {
	Tag TypeTag

	// Struct & union
	Fields []TypeHash
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
)

var table = make(map[TypeHash]TypeDummy)

func GetBuiltin(tag TypeTag) TypeHash {
	hash := TypeHash(hashU32(uint32(tag)))

	if _, ok := table[hash]; !ok {
		table[hash] = TypeDummy{Tag: tag}
	}

	return hash
}

func Register(dummy TypeDummy) TypeHash {
	hash := hashType(dummy)

	if _, ok := table[hash]; !ok {
		table[hash] = dummy
	}

	return hash
}

// 'Tag' > 0, so the hash algorithm should work. This will probably
// break, but will do for now.
func hashType(dummy TypeDummy) TypeHash {
	tagHash := TypeHash(hashU32(uint32(dummy.Tag)))

	switch dummy.Tag {
	case None:
		return tagHash

	case S64:
		return tagHash

	case U64:
		return tagHash

	case Bool:
		return tagHash

	case Struct:
		hash := TypeHash(hashU32(uint32(dummy.Tag)))
		for _, field := range dummy.Fields {
			hash = hashType(table[field])
		}
		return hash

	default:
		panic("not implemented")
	}
}

func hashU32(n uint32) uint32 {
	const knuth uint32 = 2654435769
	return n * knuth
}
