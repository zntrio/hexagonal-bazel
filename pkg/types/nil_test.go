package types

import "testing"

func TestIsNil(t *testing.T) {
	type args struct {
		c any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				c: nil,
			},
			want: true,
		},
		{
			name: "pointer to nil",
			args: args{
				c: (*struct{})(nil),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.args.c); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
