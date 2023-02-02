package brass

import (
	"fmt"
	"strings"
)

type Kind int

const (
	KindNil Kind = iota
	KindList
	KindBool
	KindString
	KindOctets
	KindInteger
	_ // reserved for KindNatural - future extension
)

type SExpr struct {
	kind    Kind
	integer int64
	octets  []byte
	list    []*SExpr
}

func (e *SExpr) Kind() Kind { return e.kind }

func (e *SExpr) IsNil() bool {
	return e.kind == KindNil
}

func (e *SExpr) AsList() []*SExpr {
	if e.kind != KindList {
		panic("must be KindList")
	}
	return e.list
}

func (e *SExpr) AsBool() bool {
	if e.kind != KindBool {
		panic("must be KindBool")
	}
	return e.integer != 0
}

func (e *SExpr) AsString() string {
	kind := e.kind
	if kind != KindString {
		panic("must be KindString")
	}
	return string(e.octets)
}

func (e *SExpr) AsOctets() []byte {
	kind := e.kind
	if kind != KindOctets {
		panic("must be KindOctets")
	}
	return e.octets
}

func (e *SExpr) AsInt64() int64 {
	kind := e.kind
	if kind != KindInteger {
		panic("must be KindInteger")
	}
	return e.integer
}

func (e *SExpr) reset() {
	e.integer = 0
	e.octets = nil
	e.list = nil
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
	e.octets = []byte(str)
}

func (e *SExpr) SetOctets(octets []byte) {
	e.reset()
	e.kind = KindOctets
	e.octets = octets
}

func (e *SExpr) SetInt64B16(value int64) {
	e.reset()
	e.kind = KindInteger
	e.integer = value
}

func (e *SExpr) String() string {
	sb := strings.Builder{}
	e.appendString(&sb)
	return sb.String()
}

func (e *SExpr) appendString(sb *strings.Builder) {
	switch e.kind {
	case KindNil:
		sb.WriteString("nil")
		return
	case KindBool:
		if e.integer != 0 {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		return
	case KindInteger:
		if e.integer < 0 {
			fmt.Fprintf(sb, "-$%x", -e.integer)
		} else {
			fmt.Fprintf(sb, "$%x", e.integer)
		}
		return
	case KindOctets:
		sb.WriteByte('#')
		fmt.Fprintf(sb, "%x", len(e.octets))
		sb.WriteByte('$')
		for _, b := range e.octets {
			fmt.Fprintf(sb, "%02x", b)
		}
		return
	case KindString:
		sb.WriteByte('"')
		for _, b := range e.octets {
			if b == '\\' {
				sb.WriteString("\\\\")
			} else if b == '"' {
				sb.WriteString("\\\"")
			} else if b == '\r' {
				sb.WriteString("\\r")
			} else if b == '\n' {
				sb.WriteString("\\n")
			} else if b == '\t' {
				sb.WriteString("\\t")
			} else if b < 32 {
				fmt.Fprintf(sb, "\\x%02x", b)
			} else if b >= 128 {
				fmt.Fprintf(sb, "\\x%02x", b)
			} else {
				sb.WriteByte(b)
			}
		}
		sb.WriteByte('"')
		return
	case KindList:
		sb.WriteByte('(')
		for i, c := range e.list {
			if i > 0 {
				sb.WriteByte(' ')
			}
			c.appendString(sb)
		}
		sb.WriteByte(')')
		return
	default:
		panic(fmt.Errorf("unimplemented kind"))
	}
}
