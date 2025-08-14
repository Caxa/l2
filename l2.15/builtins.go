package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func builtinCd(args []string) int {
	var dir string
	if len(args) == 0 {
		dir, _ = os.UserHomeDir()
	} else {
		dir = expandPath(args[0])
	}
	if dir == "" {
		fmt.Fprintln(os.Stderr, "cd: no path")
		return 1
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Fprintln(os.Stderr, "cd:", err)
		return 1
	}
	return 0
}

func builtinPwd(w io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd:", err)
		return 1
	}
	fmt.Fprintln(w, dir)
	return 0
}

func builtinEcho(args []string, w io.Writer) int {
	nl := true
	if len(args) > 0 && args[0] == "-n" {
		nl = false
		args = args[1:]
	}
	io.WriteString(w, strings.Join(args, " "))
	if nl {
		io.WriteString(w, "\n")
	}
	return 0
}

func builtinKill(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kill: usage: kill [-SIGNAL] pid")
		return 1
	}
	sig := syscall.SIGTERM
	start := 0
	if strings.HasPrefix(args[0], "-") {
		n, err := strconv.Atoi(strings.TrimPrefix(args[0], "-"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: invalid signal")
			return 1
		}
		sig = syscall.Signal(n)
		start = 1
	}
	if start >= len(args) {
		fmt.Fprintln(os.Stderr, "kill: missing pid")
		return 1
	}
	pid, err := strconv.Atoi(args[start])
	if err != nil {
		fmt.Fprintln(os.Stderr, "kill: invalid pid")
		return 1
	}
	if err := syscall.Kill(pid, sig); err != nil {
		fmt.Fprintln(os.Stderr, "kill:", err)
		return 1
	}
	return 0
}

func builtinPs(w io.Writer) int {
	// Покажем полезные колонки
	cmd := exec.Command("/bin/ps", "-o", "pid,ppid,stat,tty,time,cmd")
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ps:", err)
		return 1
	}
	return 0
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}
