package brass

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	type fields struct {
		s io.ByteScanner
	}
	tests := []struct {
		name    string
		fields  fields
		wantE   *SExpr
		wantErr bool
	}{
		// lists:
		{
			name: "()",
			fields: fields{
				s: bytes.NewBuffer([]byte("()")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{},
			},
			wantErr: false,
		},
		{
			name: "(())",
			fields: fields{
				s: bytes.NewBuffer([]byte("(())")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind: KindList,
						list: []*SExpr{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(()()())",
			fields: fields{
				s: bytes.NewBuffer([]byte("(()()())")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind: KindList,
						list: []*SExpr{},
					},
					{
						kind: KindList,
						list: []*SExpr{},
					},
					{
						kind: KindList,
						list: []*SExpr{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(",
			fields: fields{
				s: bytes.NewBuffer([]byte("(")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{},
			},
			wantErr: true,
		},
		{
			name: ")",
			fields: fields{
				s: bytes.NewBuffer([]byte("(")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{},
			},
			wantErr: true,
		},
		// integers:
		{
			name: "(1023 $3ff -1024 +$3ff +1023)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(1023 $3ff -1024 +$3ff +1023)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindInt64B10,
						integer: 1023,
					},
					{
						kind:    KindInt64B16,
						integer: 1023,
					},
					{
						kind:    KindInt64B10,
						integer: -1024,
					},
					{
						kind:    KindUInt64B16,
						integer: 1023,
					},
					{
						kind:    KindUInt64B10,
						integer: 1023,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(+$ffffffffffffffff)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(+$ffffffffffffffff)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindUInt64B16,
						integer: -1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(+$ffffff +$7e0010 +$0010)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(+$ffffff +$7e0010 +$0010)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindUInt64B16,
						integer: 0xffffff,
					},
					{
						kind:    KindUInt64B16,
						integer: 0x7e0010,
					},
					{
						kind:    KindUInt64B16,
						integer: 0x0010,
					},
				},
			},
			wantErr: false,
		},
		// tokens:
		{
			name: "(a)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(a)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsToken,
						octets: []byte("a"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(a1)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(a1)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsToken,
						octets: []byte("a1"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(b-c-d)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(b-c-d)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsToken,
						octets: []byte("b-c-d"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(. _ / ? !)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(. _ / ? !)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsToken,
						octets: []byte("."),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("_"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("/"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("?"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("!"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(a/b c_d e? f! g.h)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(a/b c_d e? f! g.h)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsToken,
						octets: []byte("a/b"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("c_d"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("e?"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("f!"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("g.h"),
					},
				},
			},
			wantErr: false,
		},
		// nil
		{
			name: "(nil @nil)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(nil @nil)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind: KindNil,
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("nil"),
					},
				},
			},
			wantErr: false,
		},
		// bool
		{
			name: "(true false @true @false)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(true false @true @false)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindBool,
						integer: -1,
					},
					{
						kind:    KindBool,
						integer: 0,
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("true"),
					},
					{
						kind:   KindOctetsToken,
						octets: []byte("false"),
					},
				},
			},
			wantErr: false,
		},
		// hex-octets
		{
			name: "(#1$00)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(#1$00)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsHex,
						octets: []byte("\x00"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(#5$fffefdfcfb)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(#5$fffefdfcfb)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsHex,
						octets: []byte("\xff\xfe\xfd\xfc\xfb"),
					},
				},
			},
			wantErr: false,
		},
		// quoted-octets
		{
			name: `("")`,
			fields: fields{
				s: bytes.NewBuffer([]byte(`("")`)),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsQuoted,
						octets: nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name: `("abc\r\n\t\xff123\\[]\"x")`,
			fields: fields{
				s: bytes.NewBuffer([]byte(`("abc\r\n\t\xff123\\[]\"x")`)),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindOctetsQuoted,
						octets: []byte("abc\r\n\t\xff123\\[]\"x"),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decoder{
				s: tt.fields.s,
			}
			gotE, err := d.Decode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotE, tt.wantE) {
				t.Errorf("Decode() gotE = %v, want %v", gotE, tt.wantE)
			}
		})
	}
}
