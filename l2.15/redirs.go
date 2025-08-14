package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type redirs struct {
	inFile  string
	outFile string
	append  bool
	hasIn   bool
	hasOut  bool
}

func parseRedirs(tokens []string) (argv []string, r redirs, err error) {
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch t {
		case "<":
			if i+1 >= len(tokens) {
				return nil, r, fmt.Errorf("syntax error: < requires file")
			}
			r.inFile = tokens[i+1]
			r.hasIn = true
			i++
		case ">":
			if i+1 >= len(tokens) {
				return nil, r, fmt.Errorf("syntax error: > requires file")
			}
			r.outFile = tokens[i+1]
			r.hasOut = true
			r.append = false
			i++
		case ">>":
			if i+1 >= len(tokens) {
				return nil, r, fmt.Errorf("syntax error: >> requires file")
			}
			r.outFile = tokens[i+1]
			r.hasOut = true
			r.append = true
			i++
		default:
			argv = append(argv, t)
		}
	}
	return
}

func setupRedirects(cmd *exec.Cmd, r redirs, closers *[]io.Closer) {
	// stdin
	if r.hasIn {
		f, err := os.Open(r.inFile)
		if err == nil {
			cmd.Stdin = f
			*closers = append(*closers, f)
		} else {
			// оставим по умолчанию os.Stdin, а ошибка всплывёт при запуске
		}
	} else if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}

	// stdout
	if r.hasOut {
		var f *os.File
		var err error
		if r.append {
			f, err = os.OpenFile(r.outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			f, err = os.OpenFile(r.outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		}
		if err == nil {
			cmd.Stdout = f
			*closers = append(*closers, f)
		}
	} else if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}

	// stderr по умолчанию — в stderr шелла
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
}

func withStdoutRedirect(r redirs, fn func(io.Writer) int) int {
	var out io.Writer = os.Stdout
	var f *os.File
	var err error

	if r.hasOut {
		if r.append {
			f, err = os.OpenFile(r.outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			f, err = os.OpenFile(r.outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "redirect open failed:", err)
			return 1
		}
		defer f.Close()
		out = f
	}
	return fn(out)
}
