package toml_test

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestMetaDataKeyInfo(t *testing.T) {
	doc := `title = "TOML Config"
count = 42
pi = 3.14
active = true
timestamp = 1979-05-27T07:32:00Z

[owner]
name = "Tom"
bio = "Author"

[database]
ports = [ 8000, 8001 ]
temp_targets = { cpu = 79.5, case = 72.0 }

[[servers]]
name = "alpha"

[[servers]]
name = "beta"
`

	var v map[string]any
	md, err := toml.Decode(doc, &v)
	if err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	tests := []struct {
		key      []string
		wantType string
		wantLine int
		wantCol  int
	}{
		{[]string{"title"}, "String", 1, 10},
		{[]string{"count"}, "Integer", 2, 9},
		{[]string{"pi"}, "Float", 3, 6},
		{[]string{"active"}, "Bool", 4, 10},
		{[]string{"timestamp"}, "Datetime", 5, 13},
		{[]string{"owner"}, "Hash", 7, 2},
		{[]string{"owner", "name"}, "String", 8, 9},
		{[]string{"owner", "bio"}, "String", 9, 8},
		{[]string{"database"}, "Hash", 11, 2},
		{[]string{"database", "ports"}, "Array", 12, 10},
		{[]string{"database", "temp_targets"}, "Hash", 13, 17},
		{[]string{"database", "temp_targets", "cpu"}, "Float", 13, 18},
		{[]string{"database", "temp_targets", "case"}, "Float", 13, 30},
		{[]string{"servers"}, "ArrayHash", 18, 2},
	}

	for _, tt := range tests {
		ki := md.KeyInfo(tt.key...)
		if ki.Type != tt.wantType {
			t.Errorf("key %v: got type %q, want %q", tt.key, ki.Type, tt.wantType)
		}
		if ki.Position.Line != tt.wantLine {
			t.Errorf("key %v: got line %d, want %d", tt.key, ki.Position.Line, tt.wantLine)
		}
		if ki.Position.Col != tt.wantCol {
			t.Errorf("key %v: got col %d, want %d", tt.key, ki.Position.Col, tt.wantCol)
		}

		pos := md.Position(tt.key...)
		if pos != ki.Position {
			t.Errorf("key %v: Position() = %+v, want %+v from KeyInfo()", tt.key, pos, ki.Position)
		}
	}
}

func TestMetaDataKeyInfoNotFound(t *testing.T) {
	doc := `a = 1`
	var v map[string]any
	md, err := toml.Decode(doc, &v)
	if err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	// Missing key
	ki := md.KeyInfo("nonexistent")
	if ki.Type != "" || ki.Position.Line != 0 {
		t.Errorf("expected zero KeyInfo for missing key, got %+v", ki)
	}
	pos := md.Position("nonexistent")
	if pos.Line != 0 {
		t.Errorf("expected zero Position for missing key, got %+v", pos)
	}

	// Empty key slice
	kiEmpty := md.KeyInfo()
	if kiEmpty.Type != "" || kiEmpty.Position.Line != 0 {
		t.Errorf("expected zero KeyInfo for empty key argument, got %+v", kiEmpty)
	}
	posEmpty := md.Position()
	if posEmpty.Line != 0 {
		t.Errorf("expected zero Position for empty key argument, got %+v", posEmpty)
	}

	// Zero MetaData
	var zeroMD toml.MetaData
	if kiZero := zeroMD.KeyInfo("a"); kiZero.Type != "" || kiZero.Position.Line != 0 {
		t.Errorf("expected zero KeyInfo from zero MetaData, got %+v", kiZero)
	}
	if posZero := zeroMD.Position("a"); posZero.Line != 0 {
		t.Errorf("expected zero Position from zero MetaData, got %+v", posZero)
	}
	if len(zeroMD.KeysInfo()) != 0 {
		t.Errorf("expected empty map from zero MetaData.KeysInfo(), got %+v", zeroMD.KeysInfo())
	}
}

func TestMetaDataKeysInfoMap(t *testing.T) {
	doc := `
[section]
key1 = "val1"
key2 = 100
`
	var v map[string]any
	md, err := toml.Decode(doc, &v)
	if err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	all := md.KeysInfo()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries in KeysInfo, got %d", len(all))
	}

	for _, k := range md.Keys() {
		ki, ok := all[k.String()]
		if !ok {
			t.Errorf("key %q not found in KeysInfo()", k.String())
			continue
		}
		if ki.Position.Line == 0 {
			t.Errorf("key %q has line 0", k.String())
		}
		// Confirm KeyInfo method returns matching data
		single := md.KeyInfo(k...)
		if single != ki {
			t.Errorf("KeyInfo(%v) = %+v, want %+v from KeysInfo", k, single, ki)
		}
	}
}

func TestMetaDataDottedAndQuotedKeys(t *testing.T) {
	doc := `
site."google.com".port = 443
plain.nested.val = "nested"
"" = "empty-key"
`
	var v map[string]any
	md, err := toml.Decode(doc, &v)
	if err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	// Dotted hierarchical key
	kiNested := md.KeyInfo("plain", "nested", "val")
	if kiNested.Type != "String" || kiNested.Position.Line != 3 {
		t.Errorf("plain.nested.val: got %+v", kiNested)
	}

	// Quoted subkey
	kiQuoted := md.KeyInfo("site", "google.com", "port")
	if kiQuoted.Type != "Integer" || kiQuoted.Position.Line != 2 {
		t.Errorf("site.\"google.com\".port: got %+v", kiQuoted)
	}

	// Empty key
	kiEmpty := md.KeyInfo("")
	if kiEmpty.Type != "String" || kiEmpty.Position.Line != 4 {
		t.Errorf("empty key: got %+v", kiEmpty)
	}

	// Fallback lookup using formatted string representation
	kiFallback := md.KeyInfo("plain.nested.val")
	if kiFallback.Type != "String" || kiFallback.Position.Line != 3 {
		t.Errorf("plain.nested.val fallback: got %+v", kiFallback)
	}
}

func TestMetaDataGitBlameUseCase(t *testing.T) {
	doc := `# Application Configuration
server_host = "127.0.0.1"
server_port = 8080

[database]
engine = "postgres"
max_conns = 25
timeout = "30s"

[[rate_limits]]
route = "/api/v1"
rps = 100
`
	var cfg map[string]any
	md, err := toml.Decode(doc, &cfg)
	if err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	lines := []string{
		"# Application Configuration",
		`server_host = "127.0.0.1"`,
		`server_port = 8080`,
		"",
		"[database]",
		`engine = "postgres"`,
		`max_conns = 25`,
		`timeout = "30s"`,
		"",
		"[[rate_limits]]",
		`route = "/api/v1"`,
		`rps = 100`,
	}

	keysToCheck := [][]string{
		{"server_host"},
		{"server_port"},
		{"database"},
		{"database", "engine"},
		{"database", "max_conns"},
		{"database", "timeout"},
		{"rate_limits"},
		{"rate_limits", "route"},
		{"rate_limits", "rps"},
	}

	for _, k := range keysToCheck {
		pos := md.Position(k...)
		if pos.Line < 1 || pos.Line > len(lines) {
			t.Fatalf("key %v line out of bounds: %d", k, pos.Line)
		}
		lineContent := lines[pos.Line-1]
		lastKey := k[len(k)-1]
		if !strings.Contains(lineContent, lastKey) {
			t.Errorf("key %v (line %d): expected %q to contain %q", k, pos.Line, lineContent, lastKey)
		}
	}
}

func TestMetaDataDecodeError(t *testing.T) {
	doc := `
[service]
name = "auth"
port = "not-an-integer"
`
	var target struct {
		Service struct {
			Name string `toml:"name"`
			Port int    `toml:"port"`
		} `toml:"service"`
	}

	md, err := toml.Decode(doc, &target)
	if err == nil {
		t.Fatal("expected decode error for type mismatch, got nil")
	}

	// Even on decode error, md contains key info from parse
	ki := md.KeyInfo("service", "port")
	if ki.Type != "String" {
		t.Errorf("got type %q, want String", ki.Type)
	}
	if ki.Position.Line != 4 {
		t.Errorf("got line %d, want 4", ki.Position.Line)
	}

	kiService := md.KeyInfo("service")
	if kiService.Type != "Hash" || kiService.Position.Line != 2 {
		t.Errorf("got service keyinfo %+v", kiService)
	}
}

