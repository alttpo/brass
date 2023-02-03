package brass

import (
	"fmt"
	"strings"
)

type Primitive interface {
	Kind() Kind
	IsNil() bool
	AsBool() bool
	AsInt64() int64
	AsString() string
	AsOctets() []byte
}

type MutablePrimitive interface {
	SetNil()
	SetBool(v bool)
	SetInt64(v int64)
	SetString(v string)
	SetOctets(v string)
}

type SExprPrimitive struct {
	kind    Kind
	integer int64
	octets  string
}

func (e *SExprPrimitive) reset() {
	e.integer = 0
	e.octets = ""
}

func (e *SExprPrimitive) SetNil() {
	e.reset()
	e.kind = KindNil
}

func (e *SExprPrimitive) SetBool(v bool) {
	e.reset()
	e.kind = KindBool
	if v {
		e.integer = -1
	} else {
		e.integer = 0
	}
}

func (e *SExprPrimitive) SetInt64(v int64) {
	e.reset()
	e.kind = KindInteger
	e.integer = v
}

func (e *SExprPrimitive) SetString(v string) {
	e.reset()
	e.kind = KindString
	e.octets = v
}

func (e *SExprPrimitive) SetOctets(v string) {
	e.reset()
	e.kind = KindOctets
	e.octets = v
}

func (e *SExprPrimitive) Kind() Kind { return e.kind }

func (e *SExprPrimitive) IsNil() bool {
	return e.kind == KindNil
}

func (e *SExprPrimitive) AsBool() bool {
	if e.kind != KindBool {
		panic("must be KindBool")
	}
	return e.integer != 0
}

func (e *SExprPrimitive) AsInt64() int64 {
	kind := e.kind
	if kind != KindInteger {
		panic("must be KindInteger")
	}
	return e.integer
}

func (e *SExprPrimitive) AsString() string {
	kind := e.kind
	if kind != KindString {
		panic("must be KindString")
	}
	return e.octets
}

func (e *SExprPrimitive) AsOctets() []byte {
	kind := e.kind
	if kind != KindOctets {
		panic("must be KindOctets")
	}
	return []byte(e.octets)
}

func (e *SExprPrimitive) AppendTo(sb *strings.Builder) {
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
		for _, b := range []byte(e.octets) {
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
	default:
		panic(fmt.Errorf("unimplemented kind"))
	}
}

func PrimitiveNil() SExprPrimitive { return SExprPrimitive{kind: KindNil} }
func PrimitiveBool(v bool) SExprPrimitive {
	if v {
		return SExprPrimitive{kind: KindBool, integer: -1}
	} else {
		return SExprPrimitive{kind: KindBool, integer: 0}
	}
}
func PrimitiveInt64(v int64) SExprPrimitive   { return SExprPrimitive{kind: KindBool, integer: v} }
func PrimitiveString(v string) SExprPrimitive { return SExprPrimitive{kind: KindString, octets: v} }
func PrimitiveOctets(v []byte) SExprPrimitive {
	return SExprPrimitive{kind: KindOctets, octets: string(v)}
}
