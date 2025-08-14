package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Тестируем builtins напрямую через io.Writer — без перехвата os.Stdout

func TestBuiltinEcho(t *testing.T) {
	var buf bytes.Buffer
	status := builtinEcho([]string{"hello", "world"}, &buf)
	if status != 0 {
		t.Fatalf("echo returned %d", status)
	}
	got := strings.TrimSpace(buf.String())
	if got != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", got)
	}
}

func TestBuiltinPwd(t *testing.T) {
	cwd, _ := os.Getwd()
	var buf bytes.Buffer
	status := builtinPwd(&buf)
	if status != 0 {
		t.Fatalf("pwd returned %d", status)
	}
	got := strings.TrimSpace(buf.String())
	if got != cwd {
		t.Errorf("expected '%s', got '%s'", cwd, got)
	}
}

func TestCdAndPwd(t *testing.T) {
	tmp := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)

	if builtinCd([]string{tmp}) != 0 {
		t.Fatalf("cd to temp dir failed")
	}

	var buf bytes.Buffer
	status := builtinPwd(&buf)
	if status != 0 {
		t.Fatalf("pwd after cd returned %d", status)
	}
	got := strings.TrimSpace(buf.String())
	if got != tmp {
		t.Errorf("expected '%s', got '%s'", tmp, got)
	}
}

// Лёгкие интеграционные проверки через системный bash —
// чтобы убедиться, что пайпы и логические операторы в принципе работают в среде.
// (Желаете — можно заменить на прямые вызовы runLine/runPipeline с фиктивным каналом.)

func TestPipelineExternal(t *testing.T) {
	cmd := exec.Command("bash", "-c", "echo hello | tr a-z A-Z")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "HELLO" {
		t.Errorf("expected 'HELLO', got '%s'", got)
	}
}

func TestLogicalAndExternal(t *testing.T) {
	cmd := exec.Command("bash", "-c", "true && echo ok")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("logical and failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "ok" {
		t.Errorf("expected 'ok', got '%s'", got)
	}
}

func TestLogicalOrExternal(t *testing.T) {
	cmd := exec.Command("bash", "-c", "false || echo yes")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("logical or failed: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "yes" {
		t.Errorf("expected 'yes', got '%s'", got)
	}
}
