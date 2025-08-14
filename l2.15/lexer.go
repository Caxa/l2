package main

import (
	"os"
	"strings"
)

func tokenize(cmd string) []string {
	var toks []string
	var b strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if b.Len() > 0 {
			toks = append(toks, b.String())
			b.Reset()
		}
	}

	for _, r := range cmd {
		if escaped {
			b.WriteRune(r)
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
				b.WriteRune(r)
			}
			continue
		}
		switch r {
		case ' ', '\t':
			flush()
		case '\'', '"':
			quote = r
		default:
			b.WriteRune(r)
		}
	}
	flush()

	// подстановка окружения — НЕ трогаем одинарные кавычки
	for i := range toks {
		toks[i] = expandEnvToken(toks[i])
	}
	return toks
}

func expandEnvToken(tok string) string {
	// 'single-quoted' — не трогаем
	if strings.HasPrefix(tok, "'") && strings.HasSuffix(tok, "'") && len(tok) >= 2 {
		return tok[1 : len(tok)-1]
	}
	// "double-quoted" и без кавычек — подставляем $VAR
	out := &strings.Builder{}
	escaped := false
	for i := 0; i < len(tok); i++ {
		ch := tok[i]
		if escaped {
			out.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '$' {
			// собрать имя переменной
			j := i + 1
			for j < len(tok) {
				c := tok[j]
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
					j++
				} else {
					break
				}
			}
			name := tok[i+1 : j]
			val := os.Getenv(name)
			out.WriteString(val)
			i = j - 1
			continue
		}
		out.WriteByte(ch)
	}
	return out.String()
}
