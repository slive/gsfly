/*
 * Author:slive
 * DATE:2020/8/11
 */
package common

import "sync"

type IParent interface {
	GetParent() interface{}

	SetParent(parent interface{})
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

func (p *Parent) SetParent(parent interface{}) {
	p.parent = parent
}

type IId interface {
	GetId() string

	SetId(id string)
}

type Id struct {
	Id string
}

func NewId() *Id {
	return &Id{Id: ""}
}

func (i *Id) GetId() string {
	return i.Id
}

func (i *Id) SetId(id string) {
	i.Id = id
}

type IAttact interface {
	AddAttach(key string, val interface{}) bool

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

func (b *Attact) AddAttach(key string, val interface{}) bool {
	if len(key) <= 0 || val == nil {
		return false
	}

	b.amut.Lock()
	defer b.amut.Unlock()
	b.attach[key] = val
	return true
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
