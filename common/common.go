/*
 * Author:slive
 * DATE:2020/8/11
 */
package common

import (
	"context"
	"fmt"
	"sync"
)

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

	Clear()
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

func (b *Attact) Clear() {
	// Todo
}

const Key_Trace = "trace"

type IRunContext interface {

	// SetContext 设置上下文，以便线程优雅停止
	SetContext(ctx context.Context)

	// GetContext 设置上下文，以便线程优雅停止
	GetContext() context.Context

	GetTrace() string

	AppendTrace(append string)

	AddTrace(append string)
}

type RunContext struct {
	context context.Context
}

// SetContext 设置上下文，以便线程优雅停止
func (ctx *RunContext) SetContext(c context.Context) {
	ctx.context = c
}

// GetContext 设置上下文，以便线程优雅停止
func (ctx *RunContext) GetContext() context.Context {
	return ctx.context
}

func (ctx *RunContext) GetTrace() string {
	value := ctx.context.Value(Key_Trace)
	if value != nil {
		s, ok := value.(string)
		if ok {
			return s
		}
	}
	return ""
}

// AppendTrace 新增到trace后面
func (ctx *RunContext) AppendTrace(append string) {
	c := ctx.context
	value := c.Value(Key_Trace)
	if value == nil {
		value = append
	} else {
		value = fmt.Sprintf("%v#%v", value, append)
	}
	ctx.context = context.WithValue(c, Key_Trace, value)
}

// AppendTrace 添加trace，覆盖原有的
func (ctx *RunContext) AddTrace(addStr string) {
	c := ctx.context
	ctx.context = context.WithValue(c, Key_Trace, addStr)
}

func NewDefRunContext() *RunContext {
	return NewRunContext(context.TODO())
}

func NewRunContext(parentCtx context.Context) *RunContext {
	var ctx context.Context
	value := parentCtx.Value(Key_Trace)
	if value != nil {
		ctx = context.WithValue(parentCtx, Key_Trace, value)
	} else {
		ctx = parentCtx
	}
	return &RunContext{context: ctx}
}

func NewRunContextByParent(parent interface{}) *RunContext {
	var ctx context.Context
	if parent != nil {
		runContext, ok := parent.(IRunContext)
		if ok {
			ctx = runContext.GetContext()
		}
	}
	if ctx == nil {
		ctx = context.TODO()
	}
	return NewRunContext(ctx)
}
