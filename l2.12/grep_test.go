package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runGrep(args []string, input string) (string, error) {
	cmd := exec.Command("go", append([]string{"run", "grep.go"}, args...)...)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestBasicMatch(t *testing.T) {
	out, err := runGrep([]string{"hello"}, "hello world\nbye\n")
	if err != nil {
		t.Fatalf("Error running grep: %v", err)
	}
	expected := "hello world\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestIgnoreCase(t *testing.T) {
	out, _ := runGrep([]string{"-i", "hello"}, "Hello\nbye\n")
	if strings.TrimSpace(out) != "Hello" {
		t.Errorf("Case-insensitive match failed: %q", out)
	}
}

func TestInvertMatch(t *testing.T) {
	out, _ := runGrep([]string{"-v", "hello"}, "hello\nbye\n")
	if strings.TrimSpace(out) != "bye" {
		t.Errorf("Invert match failed: %q", out)
	}
}

func TestCount(t *testing.T) {
	out, _ := runGrep([]string{"-c", "hello"}, "hello\nhello world\nbye\n")
	if strings.TrimSpace(out) != "2" {
		t.Errorf("Count failed: %q", out)
	}
}

func TestFixedString(t *testing.T) {
	out, _ := runGrep([]string{"-F", "he.lo"}, "he.lo\nhello\n")
	if strings.TrimSpace(out) != "he.lo" {
		t.Errorf("Fixed string match failed: %q", out)
	}
}

func TestLineNumbers(t *testing.T) {
	out, _ := runGrep([]string{"-n", "hello"}, "foo\nhello\nbar\n")
	expected := "2:hello\n"
	if out != expected {
		t.Errorf("Line numbers failed: %q", out)
	}
}

func TestContextFlags(t *testing.T) {
	out, _ := runGrep([]string{"-C", "1", "bar"}, "foo\nbar\nbaz\n")
	expected := "foo\nbar\nbaz\n"
	if out != expected {
		t.Errorf("Context failed: %q", out)
	}
}

func TestBeforeAfterFlags(t *testing.T) {
	out, _ := runGrep([]string{"-B", "1", "-A", "1", "bar"}, "foo\nbar\nbaz\nqux\n")
	expected := "foo\nbar\nbaz\n"
	if out != expected {
		t.Errorf("Before/After failed: %q", out)
	}
}

func TestFileInput(t *testing.T) {
	content := "alpha\nbeta\ngamma\n"
	tmp := "test_input.txt"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	defer os.Remove(tmp)

	out, _ := runGrep([]string{"beta", tmp}, "")
	if strings.TrimSpace(out) != "beta" {
		t.Errorf("File input failed: %q", out)
	}
}
