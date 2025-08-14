package main

import (
	"fmt"
	"sort"
	"strings"
)

func FindAnagrams(words []string) map[string][]string {
	groups := make(map[string][]string)
	firstOccurrence := make(map[string]string)
	seen := make(map[string]bool)

	for _, word := range words {
		lowWord := strings.ToLower(word)
		if seen[lowWord] {
			continue
		}
		seen[lowWord] = true

		sorted := sortString(lowWord)

		if _, exists := firstOccurrence[sorted]; !exists {
			firstOccurrence[sorted] = lowWord
		}

		groups[sorted] = append(groups[sorted], lowWord)
	}

	res := make(map[string][]string)
	for key, group := range groups {
		if len(group) >= 2 {
			res[firstOccurrence[key]] = group
		}
	}

	return res
}

func sortString(s string) string {
	chars := strings.Split(s, "")
	sort.Strings(chars)
	return strings.Join(chars, "")
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	anagrams := FindAnagrams(words)
	fmt.Println(anagrams)
}
