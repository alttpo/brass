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
			wantN:   list(octets([]byte("a"))),
		},
		{
			name:    "(a1)",
			nstr:    "(a1)",
			wantErr: "",
			wantN:   list(octets([]byte("a1"))),
		},
		{
			name:    "(b-c-d)",
			nstr:    "(b-c-d)",
			wantErr: "",
			wantN:   list(octets([]byte("b-c-d"))),
		},
		{
			name:    "(a/b c.1 d2 ? / . _ !)",
			nstr:    "(a/b c.1 d2 ? / . _ !)",
			wantErr: "",
			wantN: list(
				octets([]byte("a/b")),
				octets([]byte("c.1")),
				octets([]byte("d2")),
				octets([]byte("?")),
				octets([]byte("/")),
				octets([]byte(".")),
				octets([]byte("_")),
				octets([]byte("!")),
			),
		},
		{
			name:    "(nil true false)",
			nstr:    "(nil true false)",
			wantErr: "",
			wantN:   list(lua.LNil, lua.LTrue, lua.LFalse),
		},
		{
			name:    "(@nil @true @false)",
			nstr:    "(@nil @true @false)",
			wantErr: "",
			wantN: list(
				octets([]byte("nil")),
				octets([]byte("true")),
				octets([]byte("false")),
			),
		},
		{
			name:    "(a b c)",
			nstr:    "(a b c)",
			wantErr: "",
			wantN: list(
				octets([]byte("a")),
				octets([]byte("b")),
				octets([]byte("c")),
			),
		},
		{
			name:    "(1 $2 -$3 -4)",
			nstr:    "(1 $2 -$3 -4)",
			wantErr: "",
			wantN: list(
				lua.LNumber(1),
				lua.LNumber(2),
				lua.LNumber(-3),
				lua.LNumber(-4),
			),
		},
		{
			name:    "(#0$a)",
			nstr:    "(#0$a)",
			wantErr: "",
			wantN:   list(octets([]byte{}), octets([]byte{0x61})),
		},
		{
			name:    "(#1$61)",
			nstr:    "(#1$61)",
			wantErr: "",
			wantN:   list(octets([]byte{0x61})),
		},
		{
			name:    `("")`,
			nstr:    `("")`,
			wantErr: "",
			wantN:   list(octets([]byte{})),
		},
		{
			name:    `("a")`,
			nstr:    `("a")`,
			wantErr: "",
			wantN:   list(octets([]byte{0x61})),
		},
		{
			name:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			nstr:    `("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-=")`,
			wantErr: "",
			wantN:   list(octets([]byte("abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+-="))),
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
			wantN:   list(octets([]byte{0x61})),
		},
		{
			name:    `("cb\x61\r\n\tq")`,
			nstr:    `("cb\x61\r\n\tq")`,
			wantErr: "",
			wantN:   list(octets([]byte("cb\x61\r\n\tq"))),
		},
		{
			name:    "command",
			nstr:    "(if (eq hash \"0011223344\") (read wram $d80 16))",
			wantErr: "",
			wantN: list(
				octets([]byte("if")),
				list(
					octets([]byte("eq")),
					octets([]byte("hash")),
					octets([]byte("0011223344")),
				),
				list(
					octets([]byte("read")),
					octets([]byte("wram")),
					lua.LNumber(0xd80),
					lua.LNumber(16),
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
				wantN:   list(octets(large)),
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
				wantN:   list(octets(huge)),
			}
		}(),
	}

	l := lua.NewState(lua.Options{
		RegistrySize:    65536 * 4,
		RegistryMaxSize: 65536 * 4,
	})
	defer l.Close()

	// load the tests.lua file:
	var err error
	err = l.DoFile("brass.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			nstr := tt.nstr

			err = l.CallByParam(
				lua.P{
					Fn:      l.GetGlobal("brass_decode"),
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
func octets(s []byte) lua.LString {
	return lua.LString(s)
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
