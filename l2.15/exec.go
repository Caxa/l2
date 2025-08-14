package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

func runLine(line string, pgidCh chan *int) int {
	stmts := splitLogical(line)
	lastStatus := 0
	for i, s := range stmts {
		if i > 0 {
			prev := stmts[i-1]
			switch prev.conn {
			case connAnd:
				if lastStatus != 0 {
					continue
				}
			case connOr:
				if lastStatus == 0 {
					continue
				}
			}
		}
		status := runPipeline(s.pipeline, pgidCh)
		lastStatus = status
	}
	return lastStatus
}

func runPipeline(s string, pgidCh chan *int) int {
	stages := splitPipeline(s)
	// одиночная команда — без пайпов
	if len(stages) == 1 {
		return runSingle(stages[0])
	}

	var cmds []*exec.Cmd
	var closers []io.Closer

	// подготовим каждое звено
	for _, stage := range stages {
		argv, r, err := parseRedirs(tokenize(stage))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if len(argv) == 0 {
			fmt.Fprintln(os.Stderr, "empty stage")
			return 1
		}

		// внутри пайплайна builtins (кроме cd/exit) — через внешние утилиты-эквиваленты
		argv = externalForMaybeBuiltin(argv)

		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		setupRedirects(cmd, r, &closers)
		cmds = append(cmds, cmd)
	}

	// соединяем пайпы
	for i := 0; i < len(cmds)-1; i++ {
		pr, pw := io.Pipe()
		if cmds[i].Stdout == os.Stdout || cmds[i].Stdout == nil {
			cmds[i].Stdout = pw
		} else {
			// если был > у промежуточного — перезаписываем на пайп (упрощение)
			cmds[i].Stdout = pw
		}
		if cmds[i+1].Stdin == os.Stdin || cmds[i+1].Stdin == nil {
			cmds[i+1].Stdin = pr
		} else {
			// если был < у следующего — перезаписываем на пайп (упрощение)
			cmds[i+1].Stdin = pr
		}
		// закроем pw/pr после Wait() автоматически, когда внешние стороны закроются
	}

	// запуск
	for i, c := range cmds {
		if err := c.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "start failed:", err)
			closeMany(closers)
			return 1
		}
		if i == 0 {
			pgid := c.Process.Pid
			<-pgidCh
			pgidCh <- &pgid
		}
	}

	// ожидание
	var lastErr error
	for i := len(cmds) - 1; i >= 0; i-- {
		if err := cmds[i].Wait(); err != nil {
			lastErr = err
		}
	}
	closeMany(closers)
	<-pgidCh
	pgidCh <- nil

	return exitStatus(lastErr)
}

func runSingle(stage string) int {
	argv, r, err := parseRedirs(tokenize(stage))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(argv) == 0 {
		return 0
	}

	switch argv[0] {
	case "cd":
		return builtinCd(argv[1:])
	case "pwd":
		return withStdoutRedirect(r, func(w io.Writer) int { return builtinPwd(w) })
	case "echo":
		return withStdoutRedirect(r, func(w io.Writer) int { return builtinEcho(argv[1:], w) })
	case "kill":
		return builtinKill(argv[1:])
	case "ps":
		return withStdoutRedirect(r, func(w io.Writer) int { return builtinPs(w) })
	case "exit", "quit":
		os.Exit(0)
	}

	// внешняя команда
	var closers []io.Closer
	cmd := exec.Command(argv[0], argv[1:]...)
	setupRedirects(cmd, r, &closers)
	err = cmd.Run()
	closeMany(closers)
	return exitStatus(err)
}

func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}

func closeMany(cs []io.Closer) {
	for _, c := range cs {
		_ = c.Close()
	}
}
