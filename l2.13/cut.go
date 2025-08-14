package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseFieldsSpec(spec string) map[int]bool {
	fields := make(map[int]bool)

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			startStr := strings.TrimSpace(rangeParts[0])
			endStr := strings.TrimSpace(rangeParts[1])

			start, err1 := strconv.Atoi(startStr)
			end, err2 := strconv.Atoi(endStr)

			if err1 == nil && err2 == nil && start > 0 && end >= start {
				for i := start; i <= end; i++ {
					fields[i] = true
				}
			}
		} else {
			num, err := strconv.Atoi(part)
			if err == nil && num > 0 {
				fields[num] = true
			}
		}
	}

	return fields
}

func main() {
	fieldSpec := flag.String("f", "", "fields to extract, e.g. 1,3-5")
	delimiter := flag.String("d", "\t", "delimiter (default is tab)")
	separated := flag.Bool("s", false, "only print lines with delimiter")

	flag.Parse()

	if *fieldSpec == "" {
		fmt.Fprintln(os.Stderr, "cut: you must specify a list of fields with -f")
		os.Exit(1)
	}

	fields := parseFieldsSpec(*fieldSpec)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if *separated && !strings.Contains(line, *delimiter) {
			continue
		}

		parts := strings.Split(line, *delimiter)
		var selected []string
		for i := 1; i <= len(parts); i++ {
			if fields[i] {
				selected = append(selected, parts[i-1])
			}
		}

		if len(selected) > 0 {
			fmt.Println(strings.Join(selected, *delimiter))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}
}
