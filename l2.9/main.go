package main

import (
	"errors"
	"fmt"
	"unicode"
)

// распавоука
func Unpack(example string) (string, error) {
	var result []rune
	runes := []rune(example)
	length := len(runes)

	for i := 0; i < length; i++ {
		current := runes[i]
		// два слеша
		if current == '\\' {
			if i+1 < length {
				result = append(result, runes[1+1])
				i++
				continue
			}
			return "", errors.New("некоректная строка: escape символ без следующего символ")
		}
		//цифры
		if unicode.IsDigit(current) {
			if i == 0 {
				return "", errors.New("некоректная строка: строка начинается с цифры")
			}
			count := int(current - '0')
			if count == 0 {
				return "", errors.New("некоректная строка: множитель равен 0")
			}
			lastRun := result[len(result)-1]
			for j := 1; j < count; j++ {
				result = append(result, lastRun)

			}
			continue
		}
		result = append(result, current)

	}
	return string(result), nil
}

func main() {
	examples := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		"qwe\\4\\5",
		"qwe\\45",
	}
	for _, example := range examples {
		unpacked, err := Unpack(example)
		if err != nil {
			fmt.Printf("Вход: %q\n Ошибка %v \n", example, err)
		} else {
			fmt.Printf("Вход: %q\n Выход %v \n", example, unpacked)
		}
	}
}
