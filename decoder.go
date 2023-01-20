package brass

import (
	"errors"
	"io"
)

var ErrUnexpectedCharacter = errors.New("unexpected character")

type Decoder struct {
	s io.ByteScanner
}

func NewDecoder(s io.ByteScanner) *Decoder {
	return &Decoder{s: s}
}

func (d *Decoder) Decode() (e *SExpr, err error) {
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
			e, err = d.decodeList()
			return
		}
		if '0' <= c && c <= '9' {
			// TODO: start parsing integer
			return
		}

		err = ErrUnexpectedCharacter
		return
	}
}

func (d *Decoder) decodeList() (e *SExpr, err error) {
	var c byte
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
		child, err = d.Decode()
		if err != nil {
			return
		}
		if child == nil {
			continue
		}

		e.list = append(e.list, child)
	}
}
