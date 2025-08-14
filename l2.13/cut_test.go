package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func runCut(args []string, input string) (string, error) {
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestSingleField(t *testing.T) {
	out, _ := runCut([]string{"-f", "2"}, "a\tb\tc\n1\t2\t3\n")
	expected := "b\n2\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestMultipleFields(t *testing.T) {
	out, _ := runCut([]string{"-f", "1,3"}, "a\tb\tc\n1\t2\t3\n")
	expected := "a\tc\n1\t3\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestRange(t *testing.T) {
	out, _ := runCut([]string{"-f", "2-3"}, "a\tb\tc\td\n1\t2\t3\t4\n")
	expected := "b\tc\n2\t3\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestCustomDelimiter(t *testing.T) {
	out, _ := runCut([]string{"-f", "2", "-d", ","}, "a,b,c\n1,2,3\n")
	expected := "b\n2\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestSeparatedFlag(t *testing.T) {
	out, _ := runCut([]string{"-f", "1", "-s"}, "abc\nx\ty\n")
	expected := "x\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}
