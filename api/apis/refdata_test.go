package apis

import "testing"

func TestDataRef_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		args *DataRef
		want bool
	}{
		{
			name: "nil data",
			args: nil,
			want: true,
		},
		{
			name: "empty data",
			args: &DataRef{},
			want: true,
		},
		{
			name: "empty data",
			args: &DataRef{Name: "this is name"},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.args.IsEmpty() != test.want {
				t.Fatal()
			}
		})
	}
}
