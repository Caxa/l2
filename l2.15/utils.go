package main

import (
	"path/filepath"
	"strings"
)

func externalForMaybeBuiltin(argv []string) []string {
	if len(argv) == 0 {
		return argv
	}
	switch argv[0] {
	case "echo":
		return append([]string{"/bin/echo"}, argv[1:]...)
	case "pwd":
		return append([]string{"/bin/pwd"}, argv[1:]...)
	case "kill":
		return append([]string{"/bin/kill"}, argv[1:]...)
	case "ps":
		// добавим формат
		return append([]string{"/bin/ps", "-o", "pid,ppid,stat,tty,time,cmd"}, argv[1:]...)
	default:
		return argv
	}
}

func prependIfMissing(argv []string, path string) []string {
	if filepath.Base(argv[0]) == filepath.Base(path) && strings.HasPrefix(argv[0], "/") {
		return argv
	}
	out := make([]string, 0, len(argv)+1)
	out = append(out, path)
	out = append(out, argv[1:]...)
	return out
}
