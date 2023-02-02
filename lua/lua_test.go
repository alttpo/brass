package lua

import (
	"encoding/hex"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestLuaDecoder(t *testing.T) {
	type test struct {
		name    string
		nstr    string
		wantErr string
		wantN   lua.LValue
	}
	var cases = []test{
		{
			name:    "(a)",
			nstr:    "(a)",
			wantErr: "",
			wantN:   list(token("a")),
		},
		{
			name:    "(a1)",
			nstr:    "(a1)",
			wantErr: "",
			wantN:   list(token("a1")),
		},
		{
			name:    "(b-c-d)",
			nstr:    "(b-c-d)",
			wantErr: "",
			wantN:   list(token("b-c-d")),
		},
		{
			name:    "(a/b c.1 d2 ? / . _ !)",
			nstr:    "(a/b c.1 d2 ? / . _ !)",
			wantErr: "",
			wantN: list(
				token("a/b"),
				token("c.1"),
				token("d2"),
				token("?"),
				token("/"),
				token("."),
				token("_"),
				token("!"),
			),
		},
		{
			name:    "(nil true false)",
			nstr:    "(nil true false)",
			wantErr: "",
			wantN:   list(lnil(), lua.LTrue, lua.LFalse),
		},
		{
			name:    "(@nil @true @false)",
			nstr:    "(@nil @true @false)",
			wantErr: "",
			wantN: list(
				token("nil"),
				token("true"),
				token("false"),
			),
		},
		{
			name:    "(a b c)",
			nstr:    "(a b c)",
			wantErr: "",
			wantN: list(
				token("a"),
				token("b"),
				token("c"),
			),
		},
		{
			name:    "(1 $2 -$3 -4)",
			nstr:    "(1 $2 -$3 -4)",
			wantErr: "",
			wantN: list(
				intb10(1),
				intb16(2),
				intb16(-3),
				intb10(-4),
			),
		},
		{
			name:    "(#0$a)",
			nstr:    "(#0$a)",
			wantErr: "",
			wantN:   list(octetsHex([]byte{}), token("a")),
		},
		{
			name:    "(#1$61)",
			nstr:    "(#1$61)",
			wantErr: "",
			wantN:   list(octetsHex([]byte("a"))),
		},
		{
			name:    `("")`,
			nstr:    `("")`,
			wantErr: "",
			wantN:   list(octetsQuoted("")),
		},
		{
			name:    `("a")`,
			nstr:    `("a")`,
			wantErr: "",
			wantN:   list(octetsQuoted("a")),
		},
		{
			name:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			nstr:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantErr: "",
			wantN:   list(octetsQuoted("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")),
		},
		{
			name:    "(\"abc\ndef\")",
			nstr:    "(\"abc\ndef\")",
			wantErr: "invalid string literal",
			wantN:   list(),
		},
		{
			name:    `("\x61")`,
			nstr:    `("\x61")`,
			wantErr: "",
			wantN:   list(octetsQuoted("a")),
		},
		{
			name:    `("cb\x61\r\n\tq")`,
			nstr:    `("cb\x61\r\n\tq")`,
			wantErr: "",
			wantN:   list(octetsQuoted("cb\x61\r\n\tq")),
		},
		{
			name:    "command",
			nstr:    "(if (eq hash \"0011223344\") (read wram $d80 16))",
			wantErr: "",
			wantN: list(
				token("if"),
				list(
					token("eq"),
					token("hash"),
					octetsQuoted("0011223344"),
				),
				list(
					token("read"),
					token("wram"),
					intb16(0xd80),
					intb10(16),
				),
			),
		},
		{
			name:    "req",
			nstr:    "(v20230128/req (seq 1) (check (platform snes) (rom-hash 00112233)) (read (snes/mem/console/wram ($10 $20)) (snes/mem/cart/sram (0 $300))))",
			wantErr: "",
			wantN: list(
				token("if"),
				list(
					token("eq"),
					token("hash"),
					octetsQuoted("0011223344"),
				),
				list(
					token("read"),
					token("wram"),
					intb16(0xd80),
					intb10(16),
				),
			),
		},
		func() test {
			// fill a buffer with random bytes:
			large := make([]byte, 256)
			rand.Read(large)

			// generate a test case to match those random bytes encoded as hex-octets:
			return test{
				name:    "large hex-octets",
				nstr:    fmt.Sprintf("(#%x$%s)", len(large), hex.EncodeToString(large)),
				wantErr: "",
				wantN:   list(octetsHex(large)),
			}
		}(),
		func() test {
			// fill a buffer with random bytes:
			huge := make([]byte, 32768)
			rand.Read(huge)

			// generate a test case to match those random bytes encoded as hex-octets:
			return test{
				name:    "huge hex-octets",
				nstr:    fmt.Sprintf("(#%x$%s)", len(huge), hex.EncodeToString(huge)),
				wantErr: "",
				wantN:   list(octetsHex(huge)),
			}
		}(),
	}

	l := lua.NewState(lua.Options{
		RegistrySize:    65536 * 4,
		RegistryMaxSize: 65536 * 4,
	})
	defer l.Close()

	// load the lua file:
	var err error
	var rfn *lua.LFunction
	rfn, err = l.LoadFile("brass.lua")
	if err != nil {
		t.Fatal(err)
	}

	// call the main function to return the brass module table:
	err = l.CallByParam(lua.P{
		Fn:      rfn,
		NRet:    1,
		Protect: true,
	})

	// get the decode function out of the module:
	var br lua.LValue
	br = l.Get(-1)
	var decode *lua.LFunction
	decode = l.GetField(br, "decode").(*lua.LFunction)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			nstr := tt.nstr

			err = l.CallByParam(
				lua.P{
					Fn:      decode,
					NRet:    3,
					Protect: true,
				},
				lua.LString(nstr),
			)
			if err != nil {
				t.Fatalf("glua error: %v", err)
			}

			n, i, perr := l.Get(-3), l.Get(-2), l.Get(-1)
			l.Pop(3)

			errStr := ""
			if perr != lua.LNil {
				errStr = string(perr.(*lua.LTable).RawGetString("err").(lua.LString))
			}

			if (errStr != "") != (tt.wantErr != "") {
				t.Fatalf("want err='%v' got '%v'", tt.wantErr, errStr)
			}

			_ = i

			if !reflect.DeepEqual(tt.wantN, n) {
				t.Fatalf("want %s\ngot  %s", fmtLua(tt.wantN), fmtLua(n))
			}
		})
	}
}

func table() *lua.LTable {
	return &lua.LTable{
		Metatable: lua.LNil,
	}
}

func token(s string) lua.LString {
	return lua.LString(s)
}

func intb10(i int64) *lua.LTable {
	t := table()
	t.RawSetString("kind", lua.LString("int-b10"))
	t.RawSetString("int", lua.LNumber(i))
	return t
}

func intb16(i int64) *lua.LTable {
	t := table()
	t.RawSetString("kind", lua.LString("int-b16"))
	t.RawSetString("int", lua.LNumber(i))
	return t
}

func octetsQuoted(s string) *lua.LTable {
	t := table()
	t.RawSetString("kind", lua.LString("quoted"))
	t.RawSetString("str", lua.LString(s))
	return t
}

func octetsHex(s []byte) *lua.LTable {
	t := table()
	t.RawSetString("kind", lua.LString("hex"))
	l := table()
	for _, b := range s {
		l.Append(lua.LNumber(b))
	}
	t.RawSetString("octets", l)
	return t
}

func lnil() *lua.LTable {
	t := table()
	t.RawSetString("kind", lua.LString("nil"))
	return t
}

func list(children ...lua.LValue) *lua.LTable {
	t := table()
	for _, c := range children {
		t.Append(c)
	}
	return t
}

func fmtLua(v lua.LValue) string {
	if v == nil {
		return ""
	}

	switch v.Type() {
	case lua.LTTable:
		tb := v.(*lua.LTable)
		sb := &strings.Builder{}
		sb.WriteRune('{')
		tb.ForEach(func(key lua.LValue, val lua.LValue) {
			sb.WriteString(fmtLua(key))
			sb.WriteRune('=')
			sb.WriteString(fmtLua(val))
			sb.WriteRune(',')
		})
		if sb.Len() > 1 {
			s := sb.String()
			return s[0:len(s)-1] + "}"
		} else {
			return "{}"
		}
	case lua.LTString:
		st := string(v.(lua.LString))
		return fmt.Sprintf("%q", st)
	default:
		return v.String()
	}
}

func TestEncoder(t *testing.T) {
	type test struct {
		name    string
		e       lua.LValue
		wantErr string
		wantN   string
	}
	var cases = []test{
		{
			name:    "(a)",
			wantN:   "(a)",
			wantErr: "",
			e:       list(token("a")),
		},
		{
			name:    "(a1)",
			wantN:   "(a1)",
			wantErr: "",
			e:       list(token("a1")),
		},
		{
			name:    "(b-c-d)",
			wantN:   "(b-c-d)",
			wantErr: "",
			e:       list(token("b-c-d")),
		},
		{
			name:    "(a/b c.1 d2 ? / . _ !)",
			wantN:   "(a/b c.1 d2 ? / . _ !)",
			wantErr: "",
			e: list(
				token("a/b"),
				token("c.1"),
				token("d2"),
				token("?"),
				token("/"),
				token("."),
				token("_"),
				token("!"),
			),
		},
		{
			name:    "(nil true false)",
			wantN:   "(nil true false)",
			wantErr: "",
			e:       list(lnil(), lua.LTrue, lua.LFalse),
		},
		{
			name:    "(@nil @true @false)",
			wantN:   "(@nil @true @false)",
			wantErr: "",
			e: list(
				token("nil"),
				token("true"),
				token("false"),
			),
		},
		{
			name:    "(a b c)",
			wantN:   "(a b c)",
			wantErr: "",
			e: list(
				token("a"),
				token("b"),
				token("c"),
			),
		},
		{
			name:    "(1 $2 -$3 -4)",
			wantN:   "(1 $2 -$3 -4)",
			wantErr: "",
			e: list(
				intb10(1),
				intb16(2),
				intb16(-3),
				intb10(-4),
			),
		},
		{
			name:    "(#0$ a)",
			wantN:   "(#0$ a)",
			wantErr: "",
			e:       list(octetsHex([]byte{}), token("a")),
		},
		{
			name:    "(#1$61)",
			wantN:   "(#1$61)",
			wantErr: "",
			e:       list(octetsHex([]byte("a"))),
		},
		{
			name:    `("")`,
			wantN:   `("")`,
			wantErr: "",
			e:       list(octetsQuoted("")),
		},
		{
			name:    `("a")`,
			wantN:   `("a")`,
			wantErr: "",
			e:       list(octetsQuoted("a")),
		},
		{
			name:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantN:   `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantErr: "",
			e:       list(octetsQuoted("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")),
		},
		{
			name:    "(\"abc\\ndef\")",
			wantN:   "(\"abc\\ndef\")",
			wantErr: "",
			e:       list(octetsQuoted("abc\ndef")),
		},
		{
			name:    `("\x1f")`,
			wantN:   `("\x1f")`,
			wantErr: "",
			e:       list(octetsQuoted("\x1f")),
		},
		{
			name:    `("cba\r\n\tq")`,
			wantN:   `("cba\r\n\tq")`,
			wantErr: "",
			e:       list(octetsQuoted("cba\r\n\tq")),
		},
		{
			name:    "command",
			wantN:   "(if (eq hash \"0011223344\") (read wram $d80 16))",
			wantErr: "",
			e: list(
				token("if"),
				list(
					token("eq"),
					token("hash"),
					octetsQuoted("0011223344"),
				),
				list(
					token("read"),
					token("wram"),
					intb16(0xd80),
					intb10(16),
				),
			),
		},
	}

	l := lua.NewState(lua.Options{
		RegistrySize:    65536 * 4,
		RegistryMaxSize: 65536 * 4,
	})
	defer l.Close()

	// load the lua file:
	var err error
	var rfn *lua.LFunction
	rfn, err = l.LoadFile("brass.lua")
	if err != nil {
		t.Fatal(err)
	}

	// call the main function to return the brass module table:
	err = l.CallByParam(lua.P{
		Fn:      rfn,
		NRet:    1,
		Protect: true,
	})

	// get the encode function out of the module:
	var br lua.LValue
	br = l.Get(-1)
	var encode *lua.LFunction
	encode = l.GetField(br, "encode").(*lua.LFunction)

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err = l.CallByParam(
				lua.P{
					Fn:      encode,
					NRet:    1,
					Protect: true,
				},
				tt.e,
			)
			if err != nil {
				t.Fatalf("glua error: %v", err)
			}

			n := l.Get(-1)
			l.Pop(1)

			if !reflect.DeepEqual(tt.wantN, lua.LVAsString(n)) {
				t.Fatalf("want %s\ngot  %s", tt.wantN, lua.LVAsString(n))
			}
		})
	}
}
