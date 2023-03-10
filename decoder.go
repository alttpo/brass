package brass

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

var ErrUnexpectedCharacter = errors.New("unexpected character")
var ErrNotPrimitive = errors.New("unexpected primitive type")

type Decoder struct {
	s io.ByteScanner
}

func NewDecoder(s io.ByteScanner) *Decoder {
	return &Decoder{s: s}
}

func (d *Decoder) Decode() (e *SExpr, err error) {
	e = &SExpr{}
	err = d.decodeList(e)
	return
}

func (d *Decoder) decodeList(e *SExpr) (err error) {
	var c byte

	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c != '(' {
		err = ErrUnexpectedCharacter
		return
	}

	e.reset()
	e.kind = KindList
	e.list = make([]*SExpr, 0, 10)
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == ' ' {
			continue
		}

		if c == ')' {
			return
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		var child *SExpr = &SExpr{}
		err = d.decodeNode(child, child)
		if err != nil {
			return
		}
		if child == nil {
			continue
		}

		e.list = append(e.list, child)
	}
}

func (d *Decoder) decodeMap(e *SExpr) (err error) {
	var c byte

	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c != '{' {
		err = ErrUnexpectedCharacter
		return
	}

	e.reset()
	e.kind = KindMap
	e.dict = make(map[SExprPrimitive]*SExpr, 10)
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == ' ' {
			continue
		}

		if c == '}' {
			return
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		var key SExprPrimitive
		var value *SExpr = &SExpr{}
		err = d.decodeMapEntry(&key, value)
		if err != nil {
			return
		}

		e.dict[key] = value
	}
}

func (d *Decoder) decodeMapEntry(key MutablePrimitive, value *SExpr) (err error) {
	var c byte

	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c != '(' {
		err = ErrUnexpectedCharacter
		return
	}

	err = d.decodeNode(key, nil)
	if err != nil {
		return
	}

	err = d.decodeNode(value, value)
	if err != nil {
		return
	}

	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c != ')' {
		err = ErrUnexpectedCharacter
		return
	}

	return
}

func (d *Decoder) decodeNode(p MutablePrimitive, e *SExpr) (err error) {
	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == ' ' {
			continue
		}

		if c == ')' {
			err = d.s.UnreadByte()
			return
		}
		if c == '(' {
			if e != nil {
				err = d.s.UnreadByte()
				if err != nil {
					return
				}

				err = d.decodeList(e)
				return
			}
			err = ErrNotPrimitive
			return
		}
		if c == '{' {
			if e != nil {
				err = d.s.UnreadByte()
				if err != nil {
					return
				}

				err = d.decodeMap(e)
				return
			}
			err = ErrNotPrimitive
			return
		}

		if c == '#' {
			err = d.decodeHexOctets(p)
			return
		}
		if c == '"' {
			err = d.decodeString(p)
			return
		}
		if c == 'n' {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}
			for _, cc := range []byte("nil") {
				c, err = d.s.ReadByte()
				if err != nil {
					return
				}
				if c != cc {
					err = ErrUnexpectedCharacter
					return
				}
			}

			err = nil
			p.SetNil()
			return
		}
		if c == 't' {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}
			for _, cc := range []byte("true") {
				c, err = d.s.ReadByte()
				if err != nil {
					return
				}
				if c != cc {
					err = ErrUnexpectedCharacter
					return
				}
			}

			err = nil
			p.SetBool(true)
			return
		}
		if c == 'f' {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}
			for _, cc := range []byte("false") {
				c, err = d.s.ReadByte()
				if err != nil {
					return
				}
				if c != cc {
					err = ErrUnexpectedCharacter
					return
				}
			}

			err = nil
			p.SetBool(false)
			return
		}

		// only integer parsing beyond this point:
		negate := false
		if c == '-' {
			negate = true
			c, err = d.s.ReadByte()
			if err != nil {
				return
			}
		}

		if c == '$' {
			err = d.decodeIntB16(p, negate)
			return
		}

		err = ErrUnexpectedCharacter
		return
	}
}

func (d *Decoder) decodeIntB16(e MutablePrimitive, negate bool) (err error) {
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
		if isHexDigit(c) {
			b.WriteByte(c)
			continue
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		// signed:
		var i64 int64
		i64, err = strconv.ParseInt(b.String(), 16, 64)
		e.SetInt64(i64)
		return
	}
}

func isHexDigit(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f')
}

func (d *Decoder) decodeHexOctets(e MutablePrimitive) (err error) {
	// parse hex digits up to '$' as size:
	sizeB := bytes.Buffer{}
	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
		if c == '$' {
			break
		}
		if isHexDigit(c) {
			sizeB.WriteByte(c)
			continue
		}

		err = ErrUnexpectedCharacter
		return
	}

	// parse the first hex digits as a size of octets to parse:
	var size uint64
	size, err = strconv.ParseUint(sizeB.String(), 16, 64)
	if err != nil {
		return
	}

	// pre-allocate the buffer exactly sized:
	data := bytes.NewBuffer(make([]byte, 0, size))

	// parse hex digits as octets in pairs:
	for i := uint64(0); i < size; i++ {
		var b byte
		b, err = d.readHexByte()
		if err != nil {
			return
		}

		// append to data slice:
		data.WriteByte(b)
	}

	e.SetOctets(data.String())
	return
}

func (d *Decoder) readHexByte() (b byte, err error) {
	b = 0

	// read first digit:
	var c byte
	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if '0' <= c && c <= '9' {
		b = (c - '0') << 4
	} else if 'a' <= c && c <= 'f' {
		b = (c - 'a' + 10) << 4
	} else {
		err = ErrUnexpectedCharacter
		return
	}

	// read second digit:
	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if '0' <= c && c <= '9' {
		b |= c - '0'
	} else if 'a' <= c && c <= 'f' {
		b |= c - 'a' + 10
	} else {
		err = ErrUnexpectedCharacter
		return
	}
	return
}

func (d *Decoder) decodeString(e MutablePrimitive) (err error) {
	b := bytes.Buffer{}

	var c byte
	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}

		if c == '\r' || c == '\n' {
			err = ErrUnexpectedCharacter
			return
		}

		if c == '"' {
			break
		}

		if c == '\\' {
			c, err = d.s.ReadByte()
			if err != nil {
				return
			}

			if c == '\\' {
				b.WriteByte('\\')
			} else if c == '"' {
				b.WriteByte('"')
			} else if c == 'r' {
				b.WriteByte('\r')
			} else if c == 'n' {
				b.WriteByte('\n')
			} else if c == 't' {
				b.WriteByte('\t')
			} else if c == 'x' {
				var x byte
				x, err = d.readHexByte()
				if err != nil {
					return
				}

				b.WriteByte(x)
			}

			continue
		}

		b.WriteByte(c)
	}

	e.SetString(b.String())
	return
}
