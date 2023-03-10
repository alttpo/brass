package brass

import "testing"

func TestSExpr_String(t *testing.T) {
	type fields struct {
		kind    Kind
		integer int64
		octets  string
		list    []*SExpr
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "()",
			fields: fields{
				kind: KindList,
				list: nil,
			},
			want: "()",
		},
		{
			name: "(() ())",
			fields: fields{
				kind: KindList,
				list: []*SExpr{
					{
						kind: KindList,
					},
					{
						kind: KindList,
					},
				},
			},
			want: "(() ())",
		},
		{
			name: "(\"a\" $3 -$2)",
			fields: fields{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindString,
						octets: "a",
					},
					{
						kind:    KindInteger,
						integer: 3,
					},
					{
						kind:    KindInteger,
						integer: -2,
					},
				},
			},
			want: "(\"a\" $3 -$2)",
		},
		{
			name: `("\r\n" #3$000102)`,
			fields: fields{
				kind: KindList,
				list: []*SExpr{
					{
						kind:   KindString,
						octets: "\r\n",
					},
					{
						kind:   KindOctets,
						octets: "\x00\x01\x02",
					},
				},
			},
			want: `("\r\n" #3$000102)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &SExpr{
				kind:    tt.fields.kind,
				integer: tt.fields.integer,
				octets:  tt.fields.octets,
				list:    tt.fields.list,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
