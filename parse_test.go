package toml

import (
	"testing"
)

func TestRejectAddingKeysToNonTables(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		// Regression tests for Issue #470.
		{
			name:      "cannot add key to empty array",
			input:     "a = []\na.b = ''\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to populated array",
			input:     "a = [1, 2]\na.b = 'test'\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to empty inline table",
			input:     "a = {}\na.b = ''\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to populated inline table",
			input:     "a = { b = 1 }\na.c = ''\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to explicit table via dotted keys in super-table",
			input:     "[a.b]\n[a]\nb.c = ''\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to array of tables via dotted key",
			input:     "[[tab.arr]]\n[tab]\narr.val1 = 1\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to primitive integer",
			input:     "a = 1\na.b = 2\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to primitive string",
			input:     "a = 'hello'\na.b = 'world'\n",
			shouldErr: true,
		},
		{
			name:      "cannot add key to primitive boolean",
			input:     "a = true\na.b = false\n",
			shouldErr: true,
		},
		{
			name:      "cannot redefine dotted table with primitive",
			input:     "a.b.c = 1\na.b = 2\n",
			shouldErr: true,
		},
		// Valid cases to verify no regressions.
		{
			name:      "implicit super-table followed by explicit table definition",
			input:     "[a.b.c]\nd = 1\n[a]\ne = 2\n",
			shouldErr: false,
		},
		{
			name:      "implicit super-table defined after array of tables",
			input:     "[[a.b]]\nx = 1\n[a]\ny = 2\n",
			shouldErr: false,
		},
		{
			name:      "multiple dotted keys under same implicit table",
			input:     "a.b = 1\na.c = 2\na.d.e = 3\n",
			shouldErr: false,
		},
		{
			name:      "sub-table defined under dotted table",
			input:     "[fruit]\napple.color = 'red'\n[fruit.apple.texture]\nsmooth = true\n",
			shouldErr: false,
		},
		{
			name:      "sub-table array defined under dotted table",
			input:     "[fruit]\napple.color = 'red'\n[[fruit.apple.seeds]]\nsize = 2\n",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m map[string]any
			_, err := Decode(tt.input, &m)
			if tt.shouldErr && err == nil {
				t.Fatalf("expected error for input:\n%s\nbut got nil", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected error for input:\n%s\nerror: %v", tt.input, err)
			}
		})
	}
}
