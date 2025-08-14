package main

import (
	"strings"
	"testing"
)

func TestSortLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		config   Config
		expected []string
	}{
		{
			name:     "Базовый тест",
			input:    []string{"c", "a", "b"},
			config:   Config{},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Обратная сортировка",
			input:    []string{"a", "b", "c"},
			config:   Config{reverse: true},
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "Уникальные строки",
			input:    []string{"a", "b", "a", "c"},
			config:   Config{unique: true},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Сортировка по столбцу",
			input:    []string{"1\tb", "2\ta", "3\tc"},
			config:   Config{keyColumn: 2},
			expected: []string{"2\ta", "1\tb", "3\tc"},
		},
		{
			name:     "Числовая сортировка",
			input:    []string{"10", "2", "1"},
			config:   Config{numeric: true},
			expected: []string{"1", "2", "10"},
		},
		{
			name:     "Сортировка по месяцам",
			input:    []string{"Feb", "Jan", "Mar"},
			config:   Config{monthSort: true},
			expected: []string{"Jan", "Feb", "Mar"},
		},
		{
			name:     "Человекочитаемые размеры",
			input:    []string{"10K", "1M", "100"},
			config:   Config{humanReadable: true},
			expected: []string{"100", "10K", "1M"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortLines(tt.input, tt.config)
			if strings.Join(result, ",") != strings.Join(tt.expected, ",") {
				t.Errorf("Ожидалось: %v, получено: %v", tt.expected, result)
			}
		})
	}
}
