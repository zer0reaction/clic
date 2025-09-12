package symbol

import (
	"fmt"
)

const IdNone int = -1

type symbol struct {
	data string
}

var table []symbol

func PrintInfo(id int) {
	if id >= len(table) || id == IdNone {
		panic("invalid id")
	}
	fmt.Printf("id:%d data:%s\n", id, table[id].data)
}

func Create() int {
	id := len(table)
	table = append(table, symbol{})
	return id
}

func SetData(id int, data string) {
	if id >= len(table) || id == IdNone {
		panic("invalid id")
	}

	table[id].data = data
}
