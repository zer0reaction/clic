package symbol

import (
	"fmt"
)

const IdNone int = -1

type symbol struct {
	data string
}

var table []symbol

func PrintInfo(id int) error {
	if id >= len(table) || id == IdNone {
		return fmt.Errorf("PrintInfo: invalid id %d", id)
	}

	fmt.Printf("id:%d data:%s\n", id, table[id].data)

	return nil
}

func Create() int {
	id := len(table)
	table = append(table, symbol{})
	return id
}

func SetData(id int, data string) error {
	if id >= len(table) || id == IdNone {
		return fmt.Errorf("SetData: invalid id %d", id)
	}

	table[id].data = data
	return nil
}
