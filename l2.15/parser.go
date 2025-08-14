package main

import "strings"

type connType int

const (
	connNone connType = iota
	connAnd
	connOr
)

type stmt struct {
	pipeline string
	conn     connType // связь со следующей командой
}

func trimLine(line string) string { return strings.TrimSpace(line) }

func splitLogical(line string) []stmt {
	var res []stmt
	var b strings.Builder
	var quote rune
	escaped := false

	emit := func(ct connType) {
		s := strings.TrimSpace(b.String())
		if s != "" {
			res = append(res, stmt{pipeline: s, conn: ct})
		}
		b.Reset()
	}

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
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
		if r == '\'' || r == '"' {
			quote = r
			b.WriteRune(r)
			continue
		}
		// && и ||
		if r == '&' && i+1 < len(runes) && runes[i+1] == '&' {
			emit(connAnd)
			i++
			continue
		}
		if r == '|' && i+1 < len(runes) && runes[i+1] == '|' {
			emit(connOr)
			i++
			continue
		}
		b.WriteRune(r)
	}
	emit(connNone)
	return res
}

func splitPipeline(s string) []string {
	var parts []string
	var b strings.Builder
	var quote rune
	escaped := false

	emit := func() {
		seg := strings.TrimSpace(b.String())
		if seg != "" {
			parts = append(parts, seg)
		}
		b.Reset()
	}

	for _, r := range s {
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
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == '|' {
			emit()
			continue
		}
		b.WriteRune(r)
	}
	emit()
	return parts
}
