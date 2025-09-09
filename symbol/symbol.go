package symbol

import (
	"errors"
)

const SymbolIdNone int = -1

type symbol struct {
	data string
}

var table []symbol

func CreateSymbol() uint {
	id := uint(len(table))
	table = append(table, symbol{})
	return id
}

func SetSymbolData(id int, data string) error {
	if id >= len(table) || id == SymbolIdNone {
		return errors.New("SetSymbolData: invalid id")
	}

	table[id].data = data
	return nil
}
