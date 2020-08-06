/*
 * Author:slive
 * DATE:2020/8/6
 */
package util

type Map interface {
	Get(key interface{}) interface{}

	Add(key interface{}, val interface{})

	Remove(key interface{})

	Contain(key interface{}) bool

	Size() int

	Keys() []interface{}

	Values() []interface{}

	Entries() []*Entry

	Clear()
}

type Entry struct {
	Key interface{}

	Val interface{}
}

type BaseMap struct {
	m       map[interface{}]interface{}
	keys    []interface{}
	values  []interface{}
	entries []*Entry
}

func NewBaseMap(len int) Map {
	if len <= 0 {
		panic("len is not allowed below 0")
	}
	b := &BaseMap{
		m:       make(map[interface{}]interface{}, len),
		keys:    make([]interface{}, len),
		values:  make([]interface{}, len),
		entries: make([]*Entry, len),
	}
	return b
}

func (b *BaseMap) Get(key interface{}) interface{} {
	return b.m[key]
}

func (b *BaseMap) Add(key interface{}, val interface{}) {
	b.m[key] = val
	index := len(b.m) - 1
	b.keys[index] = key
	b.values[index] = val
	b.entries[index] = &Entry{
		Key: key,
		Val: val,
	}
}

func (b *BaseMap) Remove(key interface{}) {
	delete(b.m, key)
	for index, val := range b.values {
		if val == key {
			b.values = append(b.values[:index], b.values[index+1:]...)
			b.keys = append(b.keys[:index], b.keys[index+1:]...)
			b.entries = append(b.entries[:index], b.entries[index+1:]...)
			break
		}
	}
}

func (b *BaseMap) Size() int {
	return len(b.m)
}

func (b *BaseMap) Keys() []interface{} {
	return b.keys
}

func (b *BaseMap) Values() []interface{} {
	return b.values
}

func (b *BaseMap) Entries() []*Entry {
	return b.entries
}

func (b *BaseMap) Contain(key interface{}) bool {
	val := b.m[key]
	return val != nil
}

func (b *BaseMap) Clear() {
	values := b.values
	keys := b.keys
	entries := b.entries
	for index, key := range keys {
		delete(b.m, key)
		// TODO ??
		b.values = append(values[:index], values[index+1:]...)
		b.keys = append(keys[:index], keys[index+1:]...)
		b.entries = append(entries[:index], entries[index+1:]...)
	}
}
