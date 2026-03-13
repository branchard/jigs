package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Entry represents a single line in a .env file.
// It can be a key-value pair, a comment, or a blank line.
type Entry struct {
	Key     string // empty for comments and blank lines
	Value   string
	Comment string // raw line content for comments/blank lines
	IsVar   bool   // true if this entry is a KEY=VALUE pair
}

// File represents a parsed .env file, preserving order.
type File struct {
	Entries []Entry
}

// Parse reads a .env file and returns a File preserving structure.
func Parse(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLine(line)
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return &File{Entries: entries}, nil
}

// parseLine parses a single line from a .env file.
func parseLine(line string) Entry {
	trimmed := strings.TrimSpace(line)

	// Blank line
	if trimmed == "" {
		return Entry{Comment: line}
	}

	// Comment line
	if strings.HasPrefix(trimmed, "#") {
		return Entry{Comment: line}
	}

	// KEY=VALUE pair
	idx := strings.IndexByte(trimmed, '=')
	if idx < 0 {
		// Line without '=' — treat as comment/passthrough
		return Entry{Comment: line}
	}

	key := strings.TrimSpace(trimmed[:idx])
	value := strings.TrimSpace(trimmed[idx+1:])
	value = unquote(value)

	return Entry{Key: key, Value: value, IsVar: true}
}

// unquote removes surrounding single or double quotes from a value.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// Vars returns an ordered slice of key-value pairs for variable entries.
func (f *File) Vars() []Entry {
	var vars []Entry
	for _, e := range f.Entries {
		if e.IsVar {
			vars = append(vars, e)
		}
	}
	return vars
}

// VarMap returns a map of variable names to values.
func (f *File) VarMap() map[string]string {
	m := make(map[string]string)
	for _, e := range f.Entries {
		if e.IsVar {
			m[e.Key] = e.Value
		}
	}
	return m
}

// Set updates an existing variable's value or appends a new one.
func (f *File) Set(key, value string) {
	for i, e := range f.Entries {
		if e.IsVar && e.Key == key {
			f.Entries[i].Value = value
			return
		}
	}
	f.Entries = append(f.Entries, Entry{Key: key, Value: value, IsVar: true})
}

// Write serializes the File to the given path.
func (f *File) Write(path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	for i, e := range f.Entries {
		if e.IsVar {
			fmt.Fprintf(w, "%s=%s", e.Key, quoteIfNeeded(e.Value))
		} else {
			fmt.Fprint(w, e.Comment)
		}
		if i < len(f.Entries)-1 {
			fmt.Fprintln(w)
		}
	}
	// Always end with a newline
	fmt.Fprintln(w)
	return w.Flush()
}

// quoteIfNeeded wraps a value in double quotes if it contains spaces,
// special characters, or is empty.
func quoteIfNeeded(s string) string {
	if s == "" {
		return ""
	}
	if strings.ContainsAny(s, " \t#\"'\\$`!") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return s
}
