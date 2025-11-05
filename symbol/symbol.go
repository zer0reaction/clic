package symbol

import (
	"clic/types"
)

type Id int32

const (
	IdNone Id = -1
)

type tag uint

const (
	symbolError tag = iota
	LVar
	Fun
	Type
)

type symbol struct {
	Tag  tag
	Name string
	Type types.Id

	LVar struct {
		Offset uint
	}

	Fun struct {
		Params []TypedIdent
	}
}

type TypedIdent struct {
	Name string
	Type types.Id
}

type Table struct {
	scopeStack []map[string]Id
	data       []symbol
}

func (t *Table) PushScope() {
	t.scopeStack = append(t.scopeStack, make(map[string]Id))
}

func (t *Table) PopScope() {
	if len(t.scopeStack) == 0 {
		panic("no scopes to pop")
	}
	t.scopeStack = t.scopeStack[:1]
}

func (t *Table) Add(name string, tag tag) (Id, bool) {
	if len(t.scopeStack) == 0 {
		panic("adding symbol with no scopes")
	}

	_, ok := t.scopeStack[len(t.scopeStack)-1][name]
	if ok {
		return IdNone, false
	}

	sym := symbol{
		Name: name,
		Tag:  tag,
	}
	id := Id(len(t.data))
	t.data = append(t.data, sym)
	t.scopeStack[len(t.scopeStack)-1][name] = id

	return id, true
}

func (t *Table) Get(id Id) symbol {
	return t.data[id]
}

func (t *Table) Set(id Id, sym symbol) {
	t.data[id] = sym
}

func (t *Table) Resolve(name string) (Id, bool) {
	for i := len(t.scopeStack) - 1; i >= 0; i-- {
		id, ok := t.scopeStack[i][name]
		if ok {
			return id, true
		}
	}
	return IdNone, false
}

func (t *Table) ResolveWithTag(name string, tag tag) (Id, bool) {
	for i := len(t.scopeStack) - 1; i >= 0; i-- {
		id, ok := t.scopeStack[i][name]
		if ok && t.data[id].Tag == tag {
			return id, true
		}
	}
	return IdNone, false
}
