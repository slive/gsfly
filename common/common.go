/*
 * Author:slive
 * DATE:2020/8/11
 */
package common

import "sync"

type IParent interface {
	GetParent() interface{}
}
type Parent struct {
	parent interface{}
}

func NewParent(parent interface{}) *Parent {
	return &Parent{parent: parent}
}

func (p *Parent) GetParent() interface{} {
	return p.parent
}

type IId interface {
	GetId() string

	SetId(id string)
}

type Id struct {
	id string
}

func NewId() *Id {
	return &Id{id: ""}
}

func (i *Id) GetId() string {
	return i.id
}

func (i *Id) SetId(id string) {
	i.id = id
}

type IAttact interface {
	AddAttach(key string, val interface{})

	GetAttach(key string) interface{}

	RemoveAttach(key string)
}

type Attact struct {
	// channel存放附件，可以是任意key-value值
	attach map[string]interface{}
	amut   sync.RWMutex
}

func NewAttact() *Attact {
	a := &Attact{}
	a.attach = make(map[string]interface{})
	return a
}

func (b *Attact) AddAttach(key string, val interface{}) {
	b.amut.Lock()
	defer b.amut.Unlock()
	b.attach[key] = val
}

func (b *Attact) GetAttach(key string) interface{} {
	b.amut.RLock()
	defer b.amut.RUnlock()
	return b.attach[key]
}

func (b *Attact) RemoveAttach(key string) {
	b.amut.RLock()
	defer b.amut.RUnlock()
	delete(b.attach, key)
}
