package prompt

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestForValue_WithDefault_AcceptDefault(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("\n"))
	var output bytes.Buffer

	val, err := ForValue(input, &output, "DB_HOST", "localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "localhost" {
		t.Errorf("got %q, want %q", val, "localhost")
	}
	if !strings.Contains(output.String(), "[localhost]") {
		t.Errorf("prompt should show default, got %q", output.String())
	}
}

func TestForValue_WithDefault_Override(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("remotehost\n"))
	var output bytes.Buffer

	val, err := ForValue(input, &output, "DB_HOST", "localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "remotehost" {
		t.Errorf("got %q, want %q", val, "remotehost")
	}
}

func TestForValue_NoDefault(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("mysecret\n"))
	var output bytes.Buffer

	val, err := ForValue(input, &output, "APP_SECRET", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "mysecret" {
		t.Errorf("got %q, want %q", val, "mysecret")
	}
	if strings.Contains(output.String(), "[") {
		t.Errorf("prompt should not show brackets for empty default, got %q", output.String())
	}
}

func TestForValue_NoDefault_EmptyInput(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("\n"))
	var output bytes.Buffer

	val, err := ForValue(input, &output, "APP_SECRET", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}
