package lua

import (
	"encoding/hex"
	lua "github.com/yuin/gopher-lua"
	"math/rand"
	"testing"
)

func BenchmarkFromHex(b *testing.B) {
	l := lua.NewState(lua.Options{
		RegistrySize:    65536 * 4,
		RegistryMaxSize: 65536 * 4,
	})
	defer l.Close()

	// load the tests.lua file:
	var err error
	err = l.DoFile("bench.lua")
	if err != nil {
		b.Fatal(err)
	}

	// fill a buffer with random bytes:
	huge := make([]byte, 32768)
	rand.Read(huge)

	// generate a test case to match those random bytes encoded as hex-octets:
	nstr := hex.EncodeToString(huge)

	benchFunc := func(fn lua.LValue) func(b *testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err = l.CallByParam(
					lua.P{
						Fn:      fn,
						NRet:    3,
						Protect: false,
					},
					lua.LString(nstr),
					lua.LNumber(len(huge)),
				)
				if err != nil {
					b.Fatalf("glua error: %v", err)
				}
			}
		}
	}

	for _, fname := range []string{"fromhex_a1", "fromhex_a2", "fromhex_b", "fromhex_c", "fromhex_d", "fromhex_e"} {
		b.Run(fname, benchFunc(l.GetGlobal(fname)))
	}
}
