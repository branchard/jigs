package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles the jigs binary into the given directory and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "jigs")
	cmd := exec.Command("go", "build", "-o", bin, "../cmd/jigs")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	return bin
}

// writeFile creates a file with the given content inside dir and returns its path.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", p, err)
	}
	return p
}

// readFile returns the content of the file at the given path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

// runJigs executes the jigs binary with the given arguments and stdin input.
// workDir is used as the working directory (so .env is created there).
// It returns stdout, stderr, and any execution error.
func runJigs(t *testing.T, bin string, workDir string, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(stdin)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestBasicFlow(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\nBAR=\n")

	stdout, _, err := runJigs(t, bin, workDir, "hello\nworld\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	if !strings.Contains(stdout, "2 variable(s)") {
		t.Errorf("expected stdout to mention 2 variables, got: %s", stdout)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "FOO=hello") {
		t.Errorf("expected FOO=hello in .env, got:\n%s", content)
	}
	if !strings.Contains(content, "BAR=world") {
		t.Errorf("expected BAR=world in .env, got:\n%s", content)
	}
}

func TestAcceptDefaults(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "HOST=localhost\nPORT=5432\n")

	// Press Enter twice to accept both defaults.
	_, _, err := runJigs(t, bin, workDir, "\n\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "HOST=localhost") {
		t.Errorf("expected HOST=localhost in .env, got:\n%s", content)
	}
	if !strings.Contains(content, "PORT=5432") {
		t.Errorf("expected PORT=5432 in .env, got:\n%s", content)
	}
}

func TestOverrideDefaults(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "HOST=localhost\nPORT=5432\n")

	_, _, err := runJigs(t, bin, workDir, "remotehost\n9999\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "HOST=remotehost") {
		t.Errorf("expected HOST=remotehost in .env, got:\n%s", content)
	}
	if !strings.Contains(content, "PORT=9999") {
		t.Errorf("expected PORT=9999 in .env, got:\n%s", content)
	}
}

func TestMultipleSourceFiles(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	// First file defines FOO with a default, second file defines FOO with a
	// different default and adds BAR. First-occurrence-wins for FOO.
	writeFile(t, workDir, ".env.dist", "FOO=first\n")
	writeFile(t, workDir, ".env.dev", "FOO=second\nBAR=\n")

	// Enter to accept FOO default (should be "first"), then provide BAR.
	_, _, err := runJigs(t, bin, workDir, "\nbaz\n", ".env.dist", ".env.dev")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "FOO=first") {
		t.Errorf("expected FOO=first (first-occurrence-wins) in .env, got:\n%s", content)
	}
	if !strings.Contains(content, "BAR=baz") {
		t.Errorf("expected BAR=baz in .env, got:\n%s", content)
	}
}

func TestExistingEnvOnlyPromptsForMissing(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\nBAR=\n")
	// Pre-create .env with FOO already defined.
	writeFile(t, workDir, ".env", "FOO=existing\n")

	stdout, _, err := runJigs(t, bin, workDir, "newval\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	// Should only prompt for 1 variable (BAR).
	if !strings.Contains(stdout, "1 variable(s)") {
		t.Errorf("expected stdout to mention 1 variable, got: %s", stdout)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	// FOO should remain untouched.
	if !strings.Contains(content, "FOO=existing") {
		t.Errorf("expected FOO=existing to be preserved in .env, got:\n%s", content)
	}
	if !strings.Contains(content, "BAR=newval") {
		t.Errorf("expected BAR=newval in .env, got:\n%s", content)
	}
}

func TestAllVariablesAlreadyDefined(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\nBAR=\n")
	writeFile(t, workDir, ".env", "FOO=a\nBAR=b\n")

	stdout, _, err := runJigs(t, bin, workDir, "", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	if !strings.Contains(stdout, "All variables are already defined") {
		t.Errorf("expected 'All variables are already defined' message, got: %s", stdout)
	}

	// .env should remain unchanged.
	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "FOO=a") || !strings.Contains(content, "BAR=b") {
		t.Errorf("expected .env to remain unchanged, got:\n%s", content)
	}
}

func TestNoArguments(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	_, stderr, err := runJigs(t, bin, workDir, "")
	if err == nil {
		t.Fatal("expected jigs to fail with no arguments")
	}

	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestHelpFlag(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runJigs(t, bin, workDir, "", flag)
			if err != nil {
				t.Fatalf("jigs %s failed: %v (stderr: %s)", flag, err, stderr)
			}

			if !strings.Contains(stdout, "Usage:") {
				t.Errorf("expected usage message on stdout for %s, got: %s", flag, stdout)
			}
			if !strings.Contains(stdout, "--help") {
				t.Errorf("expected --help mentioned in output for %s, got: %s", flag, stdout)
			}
			if !strings.Contains(stdout, "--version") {
				t.Errorf("expected --version mentioned in output for %s, got: %s", flag, stdout)
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	for _, flag := range []string{"-v", "--version"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, err := runJigs(t, bin, workDir, "", flag)
			if err != nil {
				t.Fatalf("jigs %s failed: %v (stderr: %s)", flag, err, stderr)
			}

			output := strings.TrimSpace(stdout)
			if output == "" {
				t.Errorf("expected version output for %s, got empty string", flag)
			}
			// When built without ldflags, version defaults to "dev".
			if output != "dev" {
				t.Errorf("expected version 'dev' for test build, got: %s", output)
			}
		})
	}
}

func TestNonexistentSourceFile(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	_, stderr, err := runJigs(t, bin, workDir, "", "nonexistent.env")
	if err == nil {
		t.Fatal("expected jigs to fail with nonexistent source file")
	}

	if !strings.Contains(stderr, "Error reading") {
		t.Errorf("expected error message on stderr, got: %s", stderr)
	}
}

func TestEmptySourceFile(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "# only comments\n\n")

	_, stderr, err := runJigs(t, bin, workDir, "", ".env.dist")
	if err == nil {
		t.Fatal("expected jigs to fail when source has no variables")
	}

	if !strings.Contains(stderr, "No variables found") {
		t.Errorf("expected 'No variables found' on stderr, got: %s", stderr)
	}
}

func TestValuesWithSpecialCharacters(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "MY_VAR=\n")

	// Provide a value with spaces.
	_, _, err := runJigs(t, bin, workDir, "hello world\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	// The value should be present (possibly quoted).
	if !strings.Contains(content, "MY_VAR=") {
		t.Errorf("expected MY_VAR in .env, got:\n%s", content)
	}
	// Re-parse the written .env to verify the value round-trips correctly.
	// We do this by running jigs again with the same source; all vars should
	// already be defined.
	stdout, _, err := runJigs(t, bin, workDir, "", ".env.dist")
	if err != nil {
		t.Fatalf("second jigs run failed: %v", err)
	}
	if !strings.Contains(stdout, "All variables are already defined") {
		t.Errorf("expected all variables to be defined after re-run, got: %s", stdout)
	}
}

func TestIdempotency(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "A=default_a\nB=\n")

	// First run: accept default for A, provide B.
	_, _, err := runJigs(t, bin, workDir, "\nmyval\n", ".env.dist")
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	contentAfterFirst := readFile(t, filepath.Join(workDir, ".env"))

	// Second run: should report everything is defined, .env unchanged.
	stdout, _, err := runJigs(t, bin, workDir, "", ".env.dist")
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if !strings.Contains(stdout, "All variables are already defined") {
		t.Errorf("expected no prompting on second run, got: %s", stdout)
	}

	contentAfterSecond := readFile(t, filepath.Join(workDir, ".env"))
	if contentAfterFirst != contentAfterSecond {
		t.Errorf(".env changed between runs.\nAfter first:\n%s\nAfter second:\n%s", contentAfterFirst, contentAfterSecond)
	}
}

func TestOutputOptionShortFlag(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\n")

	outPath := filepath.Join(workDir, "custom.env")

	// -o custom.env
	stdout, _, err := runJigs(t, bin, workDir, "bar\n", "-o", outPath, ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	if !strings.Contains(stdout, "1 variable(s)") {
		t.Errorf("expected stdout to mention 1 variable, got: %s", stdout)
	}

	content := readFile(t, outPath)
	if !strings.Contains(content, "FOO=bar") {
		t.Errorf("expected FOO=bar in output file, got:\n%s", content)
	}

	// Default .env should NOT exist.
	if _, err := os.Stat(filepath.Join(workDir, ".env")); err == nil {
		t.Error("expected .env to not be created when -o is used")
	}
}

func TestOutputOptionLongFlag(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\n")

	outPath := filepath.Join(workDir, "custom.env")

	// --output custom.env
	_, _, err := runJigs(t, bin, workDir, "bar\n", "--output", outPath, ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, outPath)
	if !strings.Contains(content, "FOO=bar") {
		t.Errorf("expected FOO=bar in output file, got:\n%s", content)
	}
}

func TestOutputOptionLongFlagEquals(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\n")

	outPath := filepath.Join(workDir, "custom.env")

	// --output=custom.env
	_, _, err := runJigs(t, bin, workDir, "bar\n", "--output="+outPath, ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, outPath)
	if !strings.Contains(content, "FOO=bar") {
		t.Errorf("expected FOO=bar in output file, got:\n%s", content)
	}
}

func TestOutputOptionDefaultIsEnv(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\n")

	// No -o flag: should default to .env
	_, _, err := runJigs(t, bin, workDir, "bar\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "FOO=bar") {
		t.Errorf("expected FOO=bar in default .env, got:\n%s", content)
	}
}

func TestOutputOptionMissingArgument(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\n")

	for _, flag := range []string{"-o", "--output"} {
		t.Run(flag, func(t *testing.T) {
			// Pass the flag with no value after it.
			_, stderr, err := runJigs(t, bin, workDir, "", flag)
			if err == nil {
				t.Fatalf("expected jigs to fail when %s has no argument", flag)
			}

			if !strings.Contains(stderr, "requires an argument") {
				t.Errorf("expected 'requires an argument' error for %s, got: %s", flag, stderr)
			}
		})
	}
}

func TestOutputOptionExistingOutputFile(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\nBAR=\n")

	outPath := filepath.Join(workDir, "custom.env")
	writeFile(t, workDir, "custom.env", "FOO=existing\n")

	// Only BAR should be prompted.
	stdout, _, err := runJigs(t, bin, workDir, "newval\n", "-o", outPath, ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	if !strings.Contains(stdout, "1 variable(s)") {
		t.Errorf("expected stdout to mention 1 variable, got: %s", stdout)
	}

	content := readFile(t, outPath)
	if !strings.Contains(content, "FOO=existing") {
		t.Errorf("expected FOO=existing preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "BAR=newval") {
		t.Errorf("expected BAR=newval in output, got:\n%s", content)
	}
}

func TestCommentsAndBlankLinesPreservedInExistingEnv(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	writeFile(t, workDir, ".env.dist", "FOO=\nBAR=\n")
	// Pre-create .env with a comment and blank line.
	writeFile(t, workDir, ".env", "# my config\n\nFOO=keep\n")

	_, _, err := runJigs(t, bin, workDir, "added\n", ".env.dist")
	if err != nil {
		t.Fatalf("jigs failed: %v", err)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))
	if !strings.Contains(content, "# my config") {
		t.Errorf("expected comment to be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "FOO=keep") {
		t.Errorf("expected FOO=keep to be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "BAR=added") {
		t.Errorf("expected BAR=added in .env, got:\n%s", content)
	}
}

func TestWithFiles(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()

	// Copy the files from the repo into the temp directory.
	e2eDir := filepath.Join(".")

	distContent, err := os.ReadFile(filepath.Join(e2eDir, ".env.dist"))
	if err != nil {
		t.Fatalf("failed to read .env.dist: %v", err)
	}
	devContent, err := os.ReadFile(filepath.Join(e2eDir, ".env.dev"))
	if err != nil {
		t.Fatalf("failed to read .env.dev: %v", err)
	}

	writeFile(t, workDir, ".env.dist", string(distContent))
	writeFile(t, workDir, ".env.dev", string(devContent))

	// .env.dist has 36 variables, .env.dev adds 1 (DEPLOY_KEY_LOCATION).
	// Total: 37 variables prompted.
	// Variables with empty defaults that need explicit values:
	//   EMPTY (3rd var), EMPTY_SINGLE_QUOTES (5th), EMPTY_DOUBLE_QUOTES (6th),
	//   DEPLOY_KEY_LOCATION (37th).
	// All others have defaults and we press Enter to accept them (except one, to test the default override).
	//
	// Build stdin: one line per variable, in order.
	var stdinLines []string
	stdinLines = append(stdinLines, "")                 // 1. BASIC [basic] → accept default
	stdinLines = append(stdinLines, "default_override") // 2. AFTER_LINE [after_line] → override default
	stdinLines = append(stdinLines, "myempty")          // 3. EMPTY → provide value
	stdinLines = append(stdinLines, "")                 // 4. EMPTY_COMMENTS [#comments] → accept default
	stdinLines = append(stdinLines, "mysinglequote")    // 5. EMPTY_SINGLE_QUOTES → provide value
	stdinLines = append(stdinLines, "mydoublequote")    // 6. EMPTY_DOUBLE_QUOTES → provide value
	// 7-36: remaining .env.dist vars, all have defaults → accept all
	for i := 7; i <= 36; i++ {
		stdinLines = append(stdinLines, "")
	}
	stdinLines = append(stdinLines, "/tmp/deploy") // 37. DEPLOY_KEY_LOCATION → provide value

	stdin := strings.Join(stdinLines, "\n") + "\n"

	stdout, _, runErr := runJigs(t, bin, workDir, stdin, ".env.dist", ".env.dev")
	if runErr != nil {
		t.Fatalf("jigs failed: %v", runErr)
	}

	if !strings.Contains(stdout, "37 variable(s)") {
		t.Errorf("expected 37 variables prompted, got: %s", stdout)
	}

	content := readFile(t, filepath.Join(workDir, ".env"))

	// Verify the explicitly-provided values.
	checks := map[string]string{
		"AFTER_LINE":          "default_override",
		"EMPTY":               "myempty",
		"EMPTY_SINGLE_QUOTES": "mysinglequote",
		"EMPTY_DOUBLE_QUOTES": "mydoublequote",
		"DEPLOY_KEY_LOCATION": "/tmp/deploy",
	}

	for key, want := range checks {
		expected := key + "=" + want
		if !strings.Contains(content, expected) {
			t.Errorf("expected %q in .env, got:\n%s", expected, content)
		}
	}

	// Verify some default values were accepted.
	defaultChecks := map[string]string{
		"BASIC":       "basic",
		"EQUAL_SIGNS": "equals==",
		"USERNAME":    "therealnerdybeast@example.tld",
		"SPACED_KEY":  "parsed",
	}

	for key, want := range defaultChecks {
		expected := key + "=" + want
		if !strings.Contains(content, expected) {
			t.Errorf("expected %q in .env, got:\n%s", expected, content)
		}
	}
}
