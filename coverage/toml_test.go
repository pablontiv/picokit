package coverage

import (
	"strings"
	"testing"
)

func TestParseFloorsTOML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Floors
		wantErr string
	}{
		{
			name:  "default only",
			input: "default = 85\n",
			want:  Floors{Default: 85},
		},
		{
			name:  "comments and blank lines",
			input: "# policy\n\ndefault = 85 # inline\n",
			want:  Floors{Default: 85},
		},
		{
			name:  "exclude array",
			input: "default = 85\nexclude = [\"pkg/a\", \"pkg/b\"]\n",
			want:  Floors{Default: 85, Exclude: []string{"pkg/a", "pkg/b"}},
		},
		{
			name:  "packages array with trailing comma",
			input: "default = 85\npackages = [ \"pkg/a\", ]\n",
			want:  Floors{Default: 85, Packages: []string{"pkg/a"}},
		},
		{
			name:  "empty array",
			input: "default = 85\nexclude = []\n",
			want:  Floors{Default: 85, Exclude: []string{}},
		},
		{
			name:  "escaped quote and backslash in string",
			input: "default = 85\nexclude = [\"a\\\\b\", \"c\\\"d\"]\n",
			want:  Floors{Default: 85, Exclude: []string{`a\b`, `c"d`}},
		},
		{
			name:  "hash inside string is not a comment",
			input: "default = 85\nexclude = [\"pkg/#x\"]\n",
			want:  Floors{Default: 85, Exclude: []string{"pkg/#x"}},
		},
		{
			name:  "unknown key ignored",
			input: "default = 85\nfuture_knob = 12\n",
			want:  Floors{Default: 85},
		},
		{
			name:    "non-integer default",
			input:   "default = [\n",
			wantErr: "line 1",
		},
		{
			name:    "table header rejected",
			input:   "[floors]\ndefault = 85\n",
			wantErr: "line 1",
		},
		{
			name:    "unterminated string",
			input:   "default = 85\nexclude = [\"pkg/a]\n",
			wantErr: "line 2",
		},
		{
			name:    "unsupported escape",
			input:   "default = 85\nexclude = [\"a\\n\"]\n",
			wantErr: "line 2",
		},
		{
			name:    "missing equals",
			input:   "default 85\n",
			wantErr: "line 1",
		},
		{
			name:    "multi-line array rejected",
			input:   "default = 85\nexclude = [\n\"pkg/a\",\n]\n",
			wantErr: "line 2",
		},
		{
			name:    "single-quoted string rejected",
			input:   "default = 85\nexclude = ['pkg/a']\n",
			wantErr: "line 2",
		},
		{
			name:    "unicode escape rejected",
			input:   "default = 85\nexclude = [\"\\u0041\"]\n",
			wantErr: "line 2",
		},
		{
			name:    "multi-line basic string rejected",
			input:   "default = 85\nexclude = [\"\"\"a\"\"\"]\n",
			wantErr: "line 2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseFloorsTOML([]byte(tc.input))
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not contain %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFloorsTOML: %v", err)
			}
			if got.Default != tc.want.Default {
				t.Errorf("Default = %d, want %d", got.Default, tc.want.Default)
			}
			assertStrings(t, "Packages", got.Packages, tc.want.Packages)
			assertStrings(t, "Exclude", got.Exclude, tc.want.Exclude)
		})
	}
}

func assertStrings(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", field, got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", field, i, got[i], want[i])
		}
	}
}
