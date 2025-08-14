package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

const prompt = "mini$ "

func main() {
	reader := bufio.NewReader(os.Stdin)

	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt)

	pgidCh := make(chan *int, 1)
	pgidCh <- nil

	go handleInterrupts(intCh, pgidCh)

	for {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			fmt.Println()
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "read error:", err)
			continue
		}

		line = trimLine(line)
		if line == "" {
			continue
		}

		_ = runLine(line, pgidCh)
	}
}

func handleInterrupts(intCh chan os.Signal, pgidCh chan *int) {
	for range intCh {
		var current *int
		select {
		case current = <-pgidCh:
		default:
		}
		if current != nil {
			_ = syscall.Kill(-*current, syscall.SIGINT)
		} else {
			fmt.Println()
			fmt.Print(prompt)
		}
		select {
		case pgidCh <- current:
		default:
		}
	}
}
