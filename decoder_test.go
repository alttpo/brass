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
