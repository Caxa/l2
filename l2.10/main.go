package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	keyColumn     int
	numeric       bool
	reverse       bool
	unique        bool
	monthSort     bool
	trimSpace     bool
	checkOnly     bool
	humanReadable bool
}

func main() {
	config := parseFlags()
	lines := readInput()
	lines = sortLines(lines, config)
	outputLines(lines, config)
}

// parseFlags парсит флаги командной строки
func parseFlags() Config {
	var config Config
	flag.IntVar(&config.keyColumn, "k", 0, "сортировать по столбцу (колонке) №N")
	flag.BoolVar(&config.numeric, "n", false, "сортировать по числовому значению")
	flag.BoolVar(&config.reverse, "r", false, "сортировать в обратном порядке")
	flag.BoolVar(&config.unique, "u", false, "выводить только уникальные строки")
	flag.BoolVar(&config.monthSort, "M", false, "сортировать по названию месяца")
	flag.BoolVar(&config.trimSpace, "b", false, "игнорировать хвостовые пробелы")
	flag.BoolVar(&config.checkOnly, "c", false, "проверить, отсортированы ли данные")
	flag.BoolVar(&config.humanReadable, "h", false, "сортировать по человекочитаемым размерам")
	flag.Parse()
	return config
}

func readInput() []string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func sortLines(lines []string, config Config) []string {
	if config.unique {
		lines = unique(lines)
	}

	sort.SliceStable(lines, func(i, j int) bool {
		a := lines[i]
		b := lines[j]

		if config.keyColumn > 0 {
			a = getColumn(a, config.keyColumn)
			b = getColumn(b, config.keyColumn)
		}

		if config.trimSpace {
			a = strings.TrimSpace(a)
			b = strings.TrimSpace(b)
		}

		if config.numeric {
			numA, _ := strconv.Atoi(a)
			numB, _ := strconv.Atoi(b)
			return numA < numB
		}

		if config.monthSort {
			return monthToNumber(a) < monthToNumber(b)
		}

		if config.humanReadable {
			sizeA := parseHumanReadableSize(a)
			sizeB := parseHumanReadableSize(b)
			return sizeA < sizeB
		}

		return a < b
	})

	if config.reverse {
		reverseSlice(lines)
	}

	if config.checkOnly {
		if !isSorted(lines, config) {
			fmt.Println("Данные не отсортированы")
			os.Exit(1)
		}
		fmt.Println("Данные отсортированы")
		os.Exit(0)
	}

	return lines
}

func outputLines(lines []string, config Config) {
	for _, line := range lines {
		fmt.Println(line)
	}
}

func unique(lines []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

func getColumn(line string, column int) string {
	parts := strings.Split(line, "\t")
	if column <= len(parts) {
		return parts[column-1]
	}
	return ""
}

func monthToNumber(month string) int {
	months := map[string]int{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
		"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
		"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	return months[month]
}

func parseHumanReadableSize(size string) int {
	suffixes := map[string]int{
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
	}
	for suffix, multiplier := range suffixes {
		if strings.HasSuffix(size, suffix) {
			num, _ := strconv.Atoi(strings.TrimSuffix(size, suffix))
			return num * multiplier
		}
	}
	num, _ := strconv.Atoi(size)
	return num
}

func reverseSlice(lines []string) {
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
}

func isSorted(lines []string, config Config) bool {
	for i := 1; i < len(lines); i++ {
		if lines[i-1] > lines[i] {
			return false
		}
	}
	return true
}
