package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
		isVar   bool
	}{
		{"simple", "FOO=bar", "FOO", "bar", true},
		{"empty value", "FOO=", "FOO", "", true},
		{"spaces around equals", " FOO = bar ", "FOO", "bar", true},
		{"double quoted", `FOO="hello world"`, "FOO", "hello world", true},
		{"single quoted", `FOO='hello world'`, "FOO", "hello world", true},
		{"comment", "# this is a comment", "", "", false},
		{"blank line", "", "", "", false},
		{"no equals", "INVALID_LINE", "", "", false},
		{"value with hash", "FOO=bar#baz", "FOO", "bar#baz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := parseLine(tt.input)
			if entry.IsVar != tt.isVar {
				t.Errorf("IsVar = %v, want %v", entry.IsVar, tt.isVar)
			}
			if entry.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", entry.Key, tt.wantKey)
			}
			if entry.IsVar && entry.Value != tt.wantVal {
				t.Errorf("Value = %q, want %q", entry.Value, tt.wantVal)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	content := `# Database config
DB_HOST=localhost
DB_PORT=5432
DB_NAME=

# App config
APP_SECRET=
APP_DEBUG=true
`
	path := writeTempFile(t, content)

	f, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	vars := f.Vars()
	if len(vars) != 5 {
		t.Fatalf("got %d vars, want 5", len(vars))
	}

	expected := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
		"DB_NAME":    "",
		"APP_SECRET": "",
		"APP_DEBUG":  "true",
	}

	varMap := f.VarMap()
	for k, want := range expected {
		got, ok := varMap[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("VarMap[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestFileSet(t *testing.T) {
	f := &File{}
	f.Set("FOO", "bar")
	f.Set("BAZ", "qux")

	if len(f.Entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(f.Entries))
	}

	// Update existing
	f.Set("FOO", "updated")
	if f.Entries[0].Value != "updated" {
		t.Errorf("Value = %q, want %q", f.Entries[0].Value, "updated")
	}
	if len(f.Entries) != 2 {
		t.Errorf("got %d entries after update, want 2", len(f.Entries))
	}
}

func TestWriteAndReparse(t *testing.T) {
	original := `# Header comment
DB_HOST=localhost
DB_PORT=5432

# Section
APP_NAME=myapp
`
	srcPath := writeTempFile(t, original)
	f, err := Parse(srcPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	f.Set("NEW_VAR", "hello")

	outPath := filepath.Join(t.TempDir(), ".env")
	if err := f.Write(outPath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Re-parse and verify
	f2, err := Parse(outPath)
	if err != nil {
		t.Fatalf("Parse() re-read error = %v", err)
	}

	m := f2.VarMap()
	if m["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST = %q, want %q", m["DB_HOST"], "localhost")
	}
	if m["NEW_VAR"] != "hello" {
		t.Errorf("NEW_VAR = %q, want %q", m["NEW_VAR"], "hello")
	}
}

func TestQuoteIfNeeded(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"", ""},
		{"has space", `"has space"`},
		{"has#hash", `"has#hash"`},
	}
	for _, tt := range tests {
		got := quoteIfNeeded(tt.input)
		if got != tt.want {
			t.Errorf("quoteIfNeeded(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.test")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
