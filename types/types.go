// This file contains the defenition of types.
// It exists to eliminate circular dependencies.
// Treat it like a .h file.

package types

type Type uint

const (
	typeError Type = iota
	None
	S64
	U64
	Bool
)
