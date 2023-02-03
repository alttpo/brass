package brass

import (
	"fmt"
	"strings"
)

type Kind int

const (
	KindNil Kind = iota
	KindBool
	KindInteger
	_ // reserved for KindNatural - future extension
	KindString
	KindOctets
	KindList
	KindMap
)

type AppendableTo interface {
	AppendTo(sb *strings.Builder)
}

type SExpr struct {
	kind    Kind
	integer int64
	octets  string
	list    []*SExpr
	dict    map[SExprPrimitive]*SExpr
}

func (e *SExpr) Kind() Kind { return e.kind }

func (e *SExpr) IsNil() bool {
	return e.kind == KindNil
}

func (e *SExpr) AsBool() bool {
	if e.kind != KindBool {
		panic("must be KindBool")
	}
	return e.integer != 0
}

func (e *SExpr) AsInt64() int64 {
	kind := e.kind
	if kind != KindInteger {
		panic("must be KindInteger")
	}
	return e.integer
}

func (e *SExpr) AsString() string {
	kind := e.kind
	if kind != KindString {
		panic("must be KindString")
	}
	return e.octets
}

func (e *SExpr) AsOctets() []byte {
	kind := e.kind
	if kind != KindOctets {
		panic("must be KindOctets")
	}
	return []byte(e.octets)
}

func (e *SExpr) AsList() []*SExpr {
	if e.kind != KindList {
		panic("must be KindList")
	}
	return e.list
}

func (e *SExpr) AsMap() map[SExprPrimitive]*SExpr {
	if e.kind != KindMap {
		panic("must be KindMap")
	}
	return e.dict
}

func (e *SExpr) reset() {
	e.integer = 0
	e.octets = ""
	e.list = nil
	e.dict = nil
}

func (e *SExpr) SetNil() {
	e.reset()
	e.kind = KindNil
}

func (e *SExpr) SetList(list []*SExpr) {
	if list == nil {
		panic("list cannot be nil")
	}
	e.reset()
	e.kind = KindList
	e.list = list
}

func (e *SExpr) SetBool(value bool) {
	e.reset()
	e.kind = KindBool
	if value {
		e.integer = -1
	} else {
		e.integer = 0
	}
}

func (e *SExpr) SetString(str string) {
	e.reset()
	e.kind = KindString
	e.octets = str
}

func (e *SExpr) SetOctets(octets string) {
	e.reset()
	e.kind = KindOctets
	e.octets = octets
}

func (e *SExpr) SetInt64(value int64) {
	e.reset()
	e.kind = KindInteger
	e.integer = value
}

func (e *SExpr) String() string {
	sb := strings.Builder{}
	e.AppendTo(&sb)
	return sb.String()
}

func (e *SExpr) AppendTo(sb *strings.Builder) {
	switch e.kind {
	case KindNil, KindBool, KindInteger, KindOctets, KindString:
		(&SExprPrimitive{
			kind:    e.kind,
			integer: e.integer,
			octets:  e.octets,
		}).AppendTo(sb)
		return
	case KindList:
		sb.WriteByte('(')
		for i, c := range e.list {
			if i > 0 {
				sb.WriteByte(' ')
			}
			c.AppendTo(sb)
		}
		sb.WriteByte(')')
		return
	case KindMap:
		sb.WriteByte('{')
		addSpace := false
		for k, v := range e.dict {
			if addSpace {
				sb.WriteByte(' ')
			}
			addSpace = true

			sb.WriteByte('(')
			k.AppendTo(sb)
			sb.WriteByte(' ')
			v.AppendTo(sb)
			sb.WriteByte(')')
		}
		sb.WriteByte('}')
		return
	default:
		panic(fmt.Errorf("unimplemented kind"))
	}
}
