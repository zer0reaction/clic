package symbol

import (
	"fmt"
)

type symbol struct {
	id uint // not an index!

	variable struct {
		blockId uint
		name    string
		offset  uint // subtracted from RBP
	}
}

const VariableBytesize = 8

var table []symbol

func IsVariableInBlock(name string, blockId uint) bool {
	for i := 0; i < len(table); i++ {
		nameMatch := (table[i].variable.name == name)
		blockMatch := (table[i].variable.blockId == blockId)
		if nameMatch && blockMatch {
			return true
		}
	}
	return false
}

func AddVariable(name string, blockId uint) uint {
	if IsVariableInBlock(name, blockId) {
		panic("variable already declared")
	}

	var s symbol
	id := uint(len(table)) + 1

	s.id = id
	s.variable.name = name
	s.variable.blockId = blockId
	table = append(table, s)

	return id
}

func VariableSetOffset(id uint, offset uint) {
	for i := 0; i < len(table); i++ {
		if table[i].id == id {
			table[i].variable.offset = offset
			return
		}
	}
	panic(fmt.Sprintf("set offset failed, id: %d", id))
}

func VariableGetOffset(id uint) uint {
	for i := 0; i < len(table); i++ {
		if table[i].id == id {
			return table[i].variable.offset
		}
	}
	panic("get offset failed")
}

func LookupVariable(name string, blockId uint) uint {
	for i := 0; i < len(table); i++ {
		nameMatch := (table[i].variable.name == name)
		blockMatch := (table[i].variable.blockId == blockId)
		if nameMatch && blockMatch {
			return table[i].id
		}
	}
	panic("lookup failed")
}
