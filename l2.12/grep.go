package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	after := flag.Int("A", 0, "Print N lines of trailing context after matching lines")
	before := flag.Int("B", 0, "Print N lines of leading context before matching lines")
	context := flag.Int("C", 0, "Print N lines of output context")
	count := flag.Bool("c", false, "Print only a count of matching lines")
	ignoreCase := flag.Bool("i", false, "Ignore case distinctions")
	invert := flag.Bool("v", false, "Select non-matching lines")
	fixed := flag.Bool("F", false, "Interpret PATTERN as a fixed string")
	number := flag.Bool("n", false, "Prefix each line with line number")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: grep [OPTIONS] PATTERN [FILE]")
		os.Exit(1)
	}

	pattern := args[0]
	var filename string
	if len(args) > 1 {
		filename = args[1]
	}

	// -C N эквивалентно -A N -B N
	if *context > 0 {
		*after = *context
		*before = *context
	}

	// Подготовка поиска
	if *ignoreCase {
		pattern = "(?i)" + pattern
	}

	if *fixed {
		pattern = regexp.QuoteMeta(pattern)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid pattern: %v\n", err)
		os.Exit(1)
	}

	// Чтение входного потока
	var reader io.Reader
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else {
		reader = os.Stdin
	}

	scanner := bufio.NewScanner(reader)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
		os.Exit(1)
	}

	// Поиск
	matched := make([]bool, len(lines))
	for i, line := range lines {
		match := re.MatchString(line)
		if *invert {
			match = !match
		}
		if match {
			matched[i] = true
		}
	}

	// Если -c — только количество
	if *count {
		cnt := 0
		for _, m := range matched {
			if m {
				cnt++
			}
		}
		fmt.Println(cnt)
		return
	}

	// Вывод с контекстом
	printed := make([]bool, len(lines))
	for i := 0; i < len(lines); i++ {
		if matched[i] {
			start := i - *before
			if start < 0 {
				start = 0
			}
			end := i + *after
			if end >= len(lines) {
				end = len(lines) - 1
			}
			for j := start; j <= end; j++ {
				if !printed[j] {
					if *number {
						fmt.Printf("%d:%s\n", j+1, lines[j])
					} else {
						fmt.Println(lines[j])
					}
					printed[j] = true
				}
			}
		}
	}
}
