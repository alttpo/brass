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
			name: "($3ff -$3ff)",
			fields: fields{
				s: bytes.NewBuffer([]byte("($3ff -$3ff)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindInteger,
						integer: 1023,
					},
					{
						kind:    KindInteger,
						integer: -1023,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "($fffffffffffff)",
			fields: fields{
				s: bytes.NewBuffer([]byte("($fffffffffffff)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindInteger,
						integer: 0xfffffffffffff,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "($ffffff $7e0010 $0010)",
			fields: fields{
				s: bytes.NewBuffer([]byte("($ffffff $7e0010 $0010)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind:    KindInteger,
						integer: 0xffffff,
					},
					{
						kind:    KindInteger,
						integer: 0x7e0010,
					},
					{
						kind:    KindInteger,
						integer: 0x0010,
					},
				},
			},
			wantErr: false,
		},
		// nil
		{
			name: "(nil)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(nil)")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{
					{
						kind: KindNil,
					},
				},
			},
			wantErr: false,
		},
		// bool
		{
			name: "(true false)",
			fields: fields{
				s: bytes.NewBuffer([]byte("(true false)")),
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
						kind:   KindOctets,
						octets: "\x00",
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
						kind:   KindOctets,
						octets: "\xff\xfe\xfd\xfc\xfb",
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
						kind:   KindString,
						octets: "",
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
						kind:   KindString,
						octets: "abc\r\n\t\xff123\\[]\"x",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "(\"abc\ndef\")",
			fields: fields{
				s: bytes.NewBuffer([]byte("(\"abc\ndef\")")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{},
			},
			wantErr: true,
		},
		{
			name: "({(\"abc\" 1) (\"def\" 2)})",
			fields: fields{
				s: bytes.NewBuffer([]byte("({(\"abc\" 1) (\"def\" 2)})")),
			},
			wantE: &SExpr{
				kind: KindList,
				list: []*SExpr{},
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
