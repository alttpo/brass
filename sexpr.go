package brass

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
