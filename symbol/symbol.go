package symbol

import ()

type symbol struct {
	id int
}

var table []symbol

func Create() int {
	id := len(table) + 1
	table = append(table, symbol{id: id})
	return id
}
