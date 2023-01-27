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
		kind: KindList,
		list: make([]*SExpr, 0, 10),
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
		if c == '#' {
			e, err = d.decodeHexOctets()
			return
		}
		if c == '"' {
			e, err = d.decodeQuotedOctets()
			return
		}
		if c == '@' || isTokenStart(c) {
			err = d.s.UnreadByte()
			if err != nil {
				return
			}

			e, err = d.decodeToken()
			return
		}

		// only integer parsing beyond this point:
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
		if isHexDigit(c) {
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

func isHexDigit(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f')
}

func isTokenStart(c byte) bool {
	if 'a' <= c && c <= 'z' {
		return true
	}
	if 'A' <= c && c <= 'Z' {
		return true
	}
	if c == '_' || c == '.' || c == '/' || c == '?' || c == '!' {
		return true
	}
	if c >= 128 {
		return true
	}
	return false
}

func isTokenRemainder(c byte) bool {
	if 'a' <= c && c <= 'z' {
		return true
	}
	if 'A' <= c && c <= 'Z' {
		return true
	}
	if '0' <= c && c <= '9' {
		return true
	}
	if c == '-' || c == '_' || c == '.' || c == '/' || c == '?' || c == '!' {
		return true
	}
	if c >= 128 {
		return true
	}
	return false
}

func (d *Decoder) decodeToken() (e *SExpr, err error) {
	isEscaped := false
	b := bytes.Buffer{}

	var c byte
	c, err = d.s.ReadByte()
	if err != nil {
		return
	}
	if c == '@' {
		isEscaped = true
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}
	}
	if !isTokenStart(c) {
		err = ErrUnexpectedCharacter
		return
	}
	err = d.s.UnreadByte()
	if err != nil {
		return
	}

	for {
		c, err = d.s.ReadByte()
		if err != nil {
			return
		}

		if isTokenRemainder(c) {
			b.WriteByte(c)
			continue
		}

		err = d.s.UnreadByte()
		if err != nil {
			return
		}

		if !isEscaped {
			s := b.String()
			if s == "nil" {
				e = &SExpr{kind: KindNil}
				return
			} else if s == "true" {
				e = &SExpr{kind: KindBool, integer: -1}
				return
			} else if s == "false" {
				e = &SExpr{kind: KindBool, integer: 0}
				return
			}
		}
		e = &SExpr{
			kind:   KindOctetsToken,
			octets: b.Bytes(),
		}
		return
	}
}

func (d *Decoder) decodeHexOctets() (e *SExpr, err error) {
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

	e = &SExpr{
		kind:   KindOctetsHex,
		octets: data.Bytes(),
	}
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

func (d *Decoder) decodeQuotedOctets() (e *SExpr, err error) {
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

	e = &SExpr{
		kind:   KindOctetsQuoted,
		octets: b.Bytes(),
	}
	return
}
