package ui

import "testing"

func TestPrettyJSON(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		want   string
		wantOK bool
	}{
		{
			name:   "object",
			value:  `{"b":2,"a":{"d":4,"c":[1,2]}}`,
			want:   "{\n  \"b\": 2,\n  \"a\": {\n    \"d\": 4,\n    \"c\": [\n      1,\n      2\n    ]\n  }\n}",
			wantOK: true,
		},
		{
			name:   "already pretty",
			value:  "{\n  \"a\": 1\n}",
			want:   "{\n  \"a\": 1\n}",
			wantOK: true,
		},
		{
			name:   "array",
			value:  `[1,2,3]`,
			want:   "[\n  1,\n  2,\n  3\n]",
			wantOK: true,
		},
		{
			name:   "scalar",
			value:  `42`,
			want:   "42",
			wantOK: true,
		},
		{
			name:   "not json",
			value:  `hello world`,
			want:   "",
			wantOK: false,
		},
		{
			name:   "broken json",
			value:  `{"a":`,
			want:   "",
			wantOK: false,
		},
		{
			name:   "empty",
			value:  ``,
			want:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := prettyJSON([]byte(tt.value))
			if ok != tt.wantOK {
				t.Fatalf("prettyJSON(%q) ok = %v, want %v", tt.value, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("prettyJSON(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
