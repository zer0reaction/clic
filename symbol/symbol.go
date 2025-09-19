package symbol

import (
	"errors"
)

type symbol struct {
	id uint

	variable struct {
		scopeId uint
		name    string
	}
}

var table []symbol

func AddVariable(name string, scopeId uint) (uint, error) {
	for i := 0; i < len(table); i++ {
		nameMatch := table[i].variable.name == name
		scopeMatch := table[i].variable.scopeId == scopeId
		if nameMatch && scopeMatch {
			// TODO: refactor to panic
			return 0, errors.New("internal: variable already declared")
		}
	}

	var s symbol
	id := uint(len(table)) + 1

	s.id = id
	s.variable.name = name
	s.variable.scopeId = scopeId
	table = append(table, s)

	return id, nil
}
