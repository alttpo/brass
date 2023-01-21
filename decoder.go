package brass

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

var ErrUnexpectedCharacter = errors.New("unexpected character")

type Decoder struct {
	s io.ByteScanner
}

func NewDecoder(s io.ByteScanner) *Decoder {
	return &Decoder{s: s}
}

func (d *Decoder) Decode() (e *SExpr, err error) {
	return d.decodeList()
}

func (d *Decoder) decodeList() (e *SExpr, err error) {
	var c byte

	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c != '(' {
		err = ErrUnexpectedCharacter
		return
	}

	e = &SExpr{
		kind:    KindList,
		integer: 0,
		octets:  nil,
		list:    make([]*SExpr, 0, 10),
	}
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == ' ' || c == '\t' {
			continue
		}

		if c == ')' {
			return
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		var child *SExpr
		child, err = d.decodeNode()
		if err != nil {
			return
		}
		if child == nil {
			continue
		}

		e.list = append(e.list, child)
	}
}

func (d *Decoder) decodeNode() (e *SExpr, err error) {
	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == ' ' || c == '\t' {
			continue
		}

		if c == ')' {
			err = d.s.UnreadByte()
			return
		}
		if c == '(' {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}

			e, err = d.decodeList()
			return
		}

		negate := false
		unsigned := false
		if c == '-' {
			negate = true
			c, err = d.s.ReadByte()
			if err != nil {
				return
			}
		} else if c == '+' {
			unsigned = true
			c, err = d.s.ReadByte()
			if err != nil {
				return
			}
		}

		if c == '$' {
			e, err = d.decodeIntB16(negate, unsigned)
			return
		}
		if '0' <= c && c <= '9' {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}

			e, err = d.decodeIntB10(negate, unsigned)
			return
		}

		err = ErrUnexpectedCharacter
		return
	}
}

func (d *Decoder) decodeIntB10(negate bool, unsigned bool) (e *SExpr, err error) {
	b := bytes.Buffer{}
	b.Grow(17)
	if negate {
		b.WriteByte('-')
	}

	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}

		if '0' <= c && c <= '9' {
			b.WriteByte(c)
			continue
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		if unsigned {
			// unsigned:
			var u64 uint64
			u64, err = strconv.ParseUint(b.String(), 10, 64)
			e = &SExpr{
				kind:    KindUInt64B10,
				integer: int64(u64),
				octets:  nil,
				list:    nil,
			}
			return
		} else {
			// signed:
			var i64 int64
			i64, err = strconv.ParseInt(b.String(), 10, 64)
			e = &SExpr{
				kind:    KindInt64B10,
				integer: i64,
				octets:  nil,
				list:    nil,
			}
			return
		}
	}
}

func (d *Decoder) decodeIntB16(negate bool, unsigned bool) (e *SExpr, err error) {
	b := bytes.Buffer{}
	b.Grow(17)
	if negate {
		b.WriteByte('-')
	}

	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}

		// only allow hex digits:
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') {
			b.WriteByte(c)
			continue
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		if unsigned {
			// unsigned:
			var u64 uint64
			u64, err = strconv.ParseUint(b.String(), 16, 64)
			e = &SExpr{
				kind:    KindUInt64B16,
				integer: int64(u64),
				octets:  nil,
				list:    nil,
			}
			return
		} else {
			// signed:
			var i64 int64
			i64, err = strconv.ParseInt(b.String(), 16, 64)
			e = &SExpr{
				kind:    KindInt64B16,
				integer: i64,
				octets:  nil,
				list:    nil,
			}
			return
		}
	}
}
