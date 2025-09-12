package symbol

import (
	"fmt"
	"strconv"
)

const IdNone int = -1

type symbol struct {
	data         string
	integerValue int64
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

func DataToIntegerValue(id int) {
	if id >= len(table) || id == IdNone {
		panic("invalid id")
	}

	value, err := strconv.ParseInt(table[id].data, 0, 64)
	if err != nil {
		panic("incorrect integer data")
	}

	table[id].integerValue = value
}

func GetIntegerValue(id int) int64 {
	if id >= len(table) || id == IdNone {
		panic("invalid id")
	}
	return table[id].integerValue
}
