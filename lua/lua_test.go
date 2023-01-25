package lua

import (
	"fmt"
	"github.com/yuin/gopher-lua"
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
			name:    "(a b c)",
			nstr:    "(a b c)",
			wantErr: "",
			wantN: list(
				octets([]byte("a")),
				octets([]byte("b")),
				octets([]byte("c")),
			),
		},
	}

	l := lua.NewState(lua.Options{})
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
		s := sb.String()
		return s[0:len(s)-1] + "}"
	case lua.LTString:
		st := string(v.(lua.LString))
		return fmt.Sprintf("%q", st)
	default:
		return v.String()
	}
}
