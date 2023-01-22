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
	KindOctetsToken
	KindOctetsQuoted
	KindOctetsHex
	KindInt64B10
	KindInt64B16
	KindUInt64B10
	KindUInt64B16
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

func (e *SExpr) AsOctets() []byte {
	kind := e.kind
	if kind != KindOctetsToken && kind != KindOctetsHex && kind != KindOctetsQuoted {
		panic("must be KindOctetsToken | KindOctetsHex | KindOctetsQuoted")
	}
	return e.octets
}

func (e *SExpr) AsInt64() int64 {
	kind := e.kind
	if kind != KindInt64B10 && kind != KindInt64B16 {
		panic("must be KindInt64B10 | KindInt64B16")
	}
	return e.integer
}

func (e *SExpr) AsUInt64() uint64 {
	kind := e.kind
	if kind != KindUInt64B10 && kind != KindUInt64B16 {
		panic("must be KindUInt64B10 | KindUInt64B16")
	}
	return uint64(e.integer)
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

func (e *SExpr) SetOctetsToken(octets []byte) {
	e.reset()
	e.kind = KindOctetsToken
	e.octets = octets
}

func (e *SExpr) SetOctetsQuoted(octets []byte) {
	e.reset()
	e.kind = KindOctetsQuoted
	e.octets = octets
}

func (e *SExpr) SetOctetsHex(octets []byte) {
	e.reset()
	e.kind = KindOctetsHex
	e.octets = octets
}

func (e *SExpr) SetInt64B10(value int64) {
	e.reset()
	e.kind = KindInt64B10
	e.integer = value
}

func (e *SExpr) SetInt64B16(value int64) {
	e.reset()
	e.kind = KindInt64B16
	e.integer = value
}

func (e *SExpr) SetUInt64B10(value uint64) {
	e.reset()
	e.kind = KindUInt64B10
	e.integer = int64(value)
}

func (e *SExpr) SetUInt64B16(value uint64) {
	e.reset()
	e.kind = KindUInt64B16
	e.integer = int64(value)
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
	case KindInt64B10:
		fmt.Fprintf(sb, "%d", e.integer)
		return
	case KindInt64B16:
		if e.integer < 0 {
			fmt.Fprintf(sb, "-$%x", -e.integer)
		} else {
			fmt.Fprintf(sb, "$%x", e.integer)
		}
		return
	case KindUInt64B10:
		fmt.Fprintf(sb, "%d", uint64(e.integer))
		return
	case KindUInt64B16:
		fmt.Fprintf(sb, "$%x", uint64(e.integer))
		return
	case KindOctetsHex:
		sb.WriteByte('#')
		fmt.Fprintf(sb, "%x", len(e.octets))
		sb.WriteByte('$')
		for _, b := range e.octets {
			fmt.Fprintf(sb, "%02x", b)
		}
		return
	case KindOctetsQuoted:
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
	case KindOctetsToken:
		// TODO: verify characters?
		sb.Write(e.octets)
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
