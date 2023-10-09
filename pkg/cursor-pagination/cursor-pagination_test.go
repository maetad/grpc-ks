package cursorpagination

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Cursor
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				s: "Hn8DAQEGQ3Vyc29yAf-AAAEBAQZPZmZzZXQBBAAAAAX_gAEUAA==",
			},
			want: Cursor{
				Offset: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	type args struct {
		c Cursor
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				c: Cursor{
					Offset: 10,
				},
			},
			want: "Hn8DAQEGQ3Vyc29yAf-AAAEBAQZPZmZzZXQBBAAAAAX_gAEUAA==",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.args.c)
			if got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
