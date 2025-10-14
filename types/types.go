// This file contains the defenition of types.
// It exists to eliminate circular dependencies.
// Treat it like a .h file.

package types

// TODO: Add ToString() and add to error reports
type Type uint

const (
	typeError Type = iota
	None
	S64
	U64
	Bool
)
