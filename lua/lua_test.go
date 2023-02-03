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
			name:    "(nil true false)",
			nstr:    "(nil true false)",
			wantErr: "",
			wantN:   mkList(mkNil(), lua.LTrue, lua.LFalse),
		},
		{
			name:    "($1 $2 -$3 -$4)",
			nstr:    "($1 $2 -$3 -$4)",
			wantErr: "",
			wantN: mkList(
				mkInteger(1),
				mkInteger(2),
				mkInteger(-3),
				mkInteger(-4),
			),
		},
		{
			name:    "(#0$)",
			nstr:    "(#0$)",
			wantErr: "",
			wantN:   mkList(mkOctets([]byte{})),
		},
		{
			name:    "(#1$61)",
			nstr:    "(#1$61)",
			wantErr: "",
			wantN:   mkList(mkOctets([]byte("a"))),
		},
		{
			name:    `("")`,
			nstr:    `("")`,
			wantErr: "",
			wantN:   mkList(mkString("")),
		},
		{
			name:    `("a")`,
			nstr:    `("a")`,
			wantErr: "",
			wantN:   mkList(mkString("a")),
		},
		{
			name:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			nstr:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantErr: "",
			wantN:   mkList(mkString("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")),
		},
		{
			name:    "(\"abc\ndef\")",
			nstr:    "(\"abc\ndef\")",
			wantErr: "invalid string literal",
			wantN:   mkList(),
		},
		{
			name:    `("\x61")`,
			nstr:    `("\x61")`,
			wantErr: "",
			wantN:   mkList(mkString("a")),
		},
		{
			name:    `("cb\x61\r\n\tq")`,
			nstr:    `("cb\x61\r\n\tq")`,
			wantErr: "",
			wantN:   mkList(mkString("cb\x61\r\n\tq")),
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
				wantN:   mkList(mkOctets(large)),
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
				wantN:   mkList(mkOctets(huge)),
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

func mkInteger(i int64) lua.LValue {
	return lua.LNumber(i)
}

func mkString(s string) lua.LValue {
	return lua.LString(s)
}

func mkOctets(s []byte) lua.LValue {
	t := table()
	t.RawSetString("__brass_kind", lua.LString("octets"))
	for _, b := range s {
		t.Append(lua.LNumber(b))
	}
	return t
}

func mkNil() lua.LValue {
	t := table()
	t.RawSetString("__brass_kind", lua.LString("nil"))
	return t
}

func mkList(children ...lua.LValue) lua.LValue {
	t := table()
	t.RawSetString("__brass_kind", lua.LString("list"))
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
			name:    "(nil true false)",
			wantN:   "(nil true false)",
			wantErr: "",
			e:       mkList(mkNil(), lua.LTrue, lua.LFalse),
		},
		{
			name:    "($1 $2 -$3 -$4)",
			wantN:   "($1 $2 -$3 -$4)",
			wantErr: "",
			e: mkList(
				mkInteger(1),
				mkInteger(2),
				mkInteger(-3),
				mkInteger(-4),
			),
		},
		{
			name:    "(#0$ \"a\")",
			wantN:   "(#0$ \"a\")",
			wantErr: "",
			e:       mkList(mkOctets([]byte{}), mkString("a")),
		},
		{
			name:    "(#1$61)",
			wantN:   "(#1$61)",
			wantErr: "",
			e:       mkList(mkOctets([]byte("a"))),
		},
		{
			name:    `("")`,
			wantN:   `("")`,
			wantErr: "",
			e:       mkList(mkString("")),
		},
		{
			name:    `("a")`,
			wantN:   `("a")`,
			wantErr: "",
			e:       mkList(mkString("a")),
		},
		{
			name:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantN:   `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantErr: "",
			e:       mkList(mkString("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")),
		},
		{
			name:    "(\"abc\\ndef\")",
			wantN:   "(\"abc\\ndef\")",
			wantErr: "",
			e:       mkList(mkString("abc\ndef")),
		},
		{
			name:    `("\x1f")`,
			wantN:   `("\x1f")`,
			wantErr: "",
			e:       mkList(mkString("\x1f")),
		},
		{
			name:    `("cba\r\n\tq")`,
			wantN:   `("cba\r\n\tq")`,
			wantErr: "",
			e:       mkList(mkString("cba\r\n\tq")),
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
