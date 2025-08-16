//go:build unix
// +build unix

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type cmdUnit struct {
	argv    []string
	inFile  string // "< file"
	outFile string // "> file" (truncate)
}

type pipeline struct {
	cmds []cmdUnit
}

type seqItem struct {
	pl pipeline
	op string // "", "&&", "||"  (op is operator that follows this item)
}

// global state to send SIGINT to running pipeline
var currentPGID int // 0 means idle

func main() {
	fmt.Println("mini$hell (Ctrl+D = exit, Ctrl+C = interrupt job)")
	reader := bufio.NewScanner(os.Stdin)

	// handle Ctrl+C: do not exit the shell, just interrupt current job group
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func() {
		for range sigc {
			pgid := currentPGID
			if pgid != 0 {
				// send SIGINT to the whole process group
				_ = syscall.Kill(-pgid, syscall.SIGINT)
			} else {
				// nothing is running; print a fresh prompt line
				fmt.Println()
				printPrompt()
			}
		}
	}()

	printPrompt()
	for {
		if !reader.Scan() {
			// EOF (Ctrl+D)
			fmt.Println()
			return
		}
		line := reader.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			printPrompt()
			continue
		}
		// env substitution (very simple): $VAR -> os.Getenv(VAR)
		line = os.ExpandEnv(line)

		items, err := parseSequence(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			printPrompt()
			continue
		}

		lastOK := true
		for i, item := range items {
			// short-circuit for && / ||
			if i > 0 {
				prevOp := items[i-1].op
				if prevOp == "&&" && !lastOK {
					// skip
					continue
				}
				if prevOp == "||" && lastOK {
					// skip
					continue
				}
			}
			ok, err := runPipeline(item.pl)
			lastOK = ok
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}
		printPrompt()
	}
}

func printPrompt() {
	wd, _ := os.Getwd()
	fmt.Printf("%s$ ", filepath.Base(wd))
}

// ---------- Parsing ----------

func parseSequence(line string) ([]seqItem, error) {
	// insert spaces around operators | < > && ||
	replacer := strings.NewReplacer(
		"&&", " && ",
		"||", " || ",
		"|", " | ",
		"<", " < ",
		">", " > ",
	)
	line = replacer.Replace(line)

	toks := fieldsRespectQuotes(line)
	if len(toks) == 0 {
		return nil, nil
	}

	// split by && and || into pipelines
	var items []seqItem
	var cur []string
	var ops []string
	for i := 0; i < len(toks); i++ {
		if toks[i] == "&&" || toks[i] == "||" {
			if len(cur) == 0 {
				return nil, errors.New("empty command before operator")
			}
			pl, err := parsePipelineTokens(cur)
			if err != nil {
				return nil, err
			}
			items = append(items, seqItem{pl: pl, op: toks[i]})
			cur = nil
			ops = append(ops, toks[i])
		} else {
			cur = append(cur, toks[i])
		}
	}
	if len(cur) > 0 {
		pl, err := parsePipelineTokens(cur)
		if err != nil {
			return nil, err
		}
		items = append(items, seqItem{pl: pl})
	}
	return items, nil
}

func parsePipelineTokens(toks []string) (pipeline, error) {
	var cmds []cmdUnit
	var cur cmdUnit
	expectFile := "" // "in" or "out"
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch t {
		case "|":
			if len(cur.argv) == 0 {
				return pipeline{}, errors.New("empty command in pipeline")
			}
			cmds = append(cmds, cur)
			cur = cmdUnit{}
			expectFile = ""
		case "<":
			expectFile = "in"
		case ">":
			expectFile = "out"
		default:
			if expectFile == "in" {
				cur.inFile = t
				expectFile = ""
			} else if expectFile == "out" {
				cur.outFile = t
				expectFile = ""
			} else {
				cur.argv = append(cur.argv, t)
			}
		}
	}
	if len(cur.argv) > 0 || cur.inFile != "" || cur.outFile != "" {
		cmds = append(cmds, cur)
	}
	if len(cmds) == 0 {
		return pipeline{}, errors.New("empty pipeline")
	}
	return pipeline{cmds: cmds}, nil
}

// Splits by spaces but keeps quoted chunks "like this" or 'like this'
func fieldsRespectQuotes(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' || r == '\n' {
			flush()
			continue
		}
		cur.WriteRune(r)
	}
	flush()
	return out
}

// ---------- Execution ----------

func runPipeline(pl pipeline) (bool, error) {
	// Builtins are handled only when there's a single command and it is a builtin.
	// This keeps pipeline execution simpler and satisfies the task requirements.
	if len(pl.cmds) == 1 {
		if isBuiltin(pl.cmds[0].argv) {
			return runBuiltin(pl.cmds[0])
		}
	}

	// Build the pipeline of external commands
	n := len(pl.cmds)
	cmds := make([]*exec.Cmd, n)
	filesToClose := []io.Closer{}

	// Create pipes between commands
	var prevR *os.File
	for i := 0; i < n; i++ {
		cu := pl.cmds[i]
		if len(cu.argv) == 0 {
			closeMany(filesToClose)
			if prevR != nil {
				_ = prevR.Close()
			}
			return false, errors.New("empty command")
		}
		cmd := exec.Command(cu.argv[0], cu.argv[1:]...)
		// set new process group so we can send SIGINT to entire pipeline
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		// stdin
		if i == 0 {
			if cu.inFile != "" {
				f, err := os.Open(cu.inFile)
				if err != nil {
					closeMany(filesToClose)
					if prevR != nil {
						_ = prevR.Close()
					}
					return false, fmt.Errorf("open %s: %w", cu.inFile, err)
				}
				cmd.Stdin = f
				filesToClose = append(filesToClose, f)
			} else {
				cmd.Stdin = os.Stdin
			}
		} else {
			cmd.Stdin = prevR
		}

		// stdout
		if i == n-1 {
			if cu.outFile != "" {
				f, err := os.Create(cu.outFile)
				if err != nil {
					closeMany(filesToClose)
					if prevR != nil {
						_ = prevR.Close()
					}
					return false, fmt.Errorf("create %s: %w", cu.outFile, err)
				}
				cmd.Stdout = f
				filesToClose = append(filesToClose, f)
			} else {
				cmd.Stdout = os.Stdout
			}
		} else {
			pr, pw, err := os.Pipe()
			if err != nil {
				closeMany(filesToClose)
				if prevR != nil {
					_ = prevR.Close()
				}
				return false, err
			}
			cmd.Stdout = pw
			// next command will read from pr
			prevR = pr
			filesToClose = append(filesToClose, pw)
		}

		cmd.Stderr = os.Stderr
		cmds[i] = cmd
	}

	// Start all
	for i, c := range cmds {
		if err := c.Start(); err != nil {
			closeMany(filesToClose)
			return false, err
		}
		// set the global PGID from the first started process
		if i == 0 {
			pgid, _ := syscall.Getpgid(c.Process.Pid)
			currentPGID = pgid
		}
		// close writer ends in parent as soon as possible
		// (those were collected in filesToClose)
	}

	// parent must close any write-ends to avoid deadlocks
	closeMany(filesToClose)

	// Wait for all, capture last exit status
	var lastStatus int
	var waitErr error
	for _, c := range cmds {
		if err := c.Wait(); err != nil {
			waitErr = err
			if exit, ok := err.(*exec.ExitError); ok {
				if status, ok := exit.Sys().(syscall.WaitStatus); ok {
					lastStatus = status.ExitStatus()
				}
			} else {
				lastStatus = 1
			}
		} else {
			// success
			lastStatus = 0
		}
	}
	currentPGID = 0
	return lastStatus == 0, waitErr
}

func closeMany(cs []io.Closer) {
	for _, c := range cs {
		_ = c.Close()
	}
}

// ---------- Builtins ----------

func isBuiltin(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch argv[0] {
	case "cd", "pwd", "echo", "kill", "ps", "exit":
		return true
	default:
		return false
	}
}

func runBuiltin(cu cmdUnit) (bool, error) {
	argv := cu.argv
	switch argv[0] {
	case "exit":
		os.Exit(0)
	case "cd":
		path := ""
		if len(argv) > 1 {
			path = argv[1]
		} else {
			path = os.Getenv("HOME")
		}
		if path == "" {
			return false, errors.New("cd: empty path")
		}
		if err := os.Chdir(path); err != nil {
			return false, fmt.Errorf("cd: %w", err)
		}
		return true, nil
	case "pwd":
		wd, err := os.Getwd()
		if err != nil {
			return false, err
		}
		return writeToOut(wd+"\n", cu.outFile)
	case "echo":
		text := strings.Join(argv[1:], " ") + "\n"
		return writeToOut(text, cu.outFile)
	case "kill":
		if len(argv) < 2 {
			return false, errors.New("kill: pid required")
		}
		pid, err := strconv.Atoi(argv[1])
		if err != nil {
			return false, fmt.Errorf("kill: bad pid: %v", err)
		}
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			return false, fmt.Errorf("kill: %w", err)
		}
		return true, nil
	case "ps":
		// minimal cross-unix: just delegate to system ps
		cmd := exec.Command("ps", "aux")
		cmd.Stdin = os.Stdin
		if cu.outFile != "" {
			f, err := os.Create(cu.outFile)
			if err != nil {
				return false, err
			}
			defer f.Close()
			cmd.Stdout = f
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Run(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("unknown builtin: %s", argv[0])
}

func writeToOut(s string, outFile string) (bool, error) {
	if outFile == "" {
		_, err := io.WriteString(os.Stdout, s)
		return err == nil, err
	}
	f, err := os.Create(outFile)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = io.WriteString(f, s)
	return err == nil, err
}
