package main

import "strings"

type textBuffer struct {
	builders []strings.Builder
	prefix   string
	suffix   string
}

func newTextBuffer(n int, prefix, suffix string) *textBuffer {
	buffer := &textBuffer{
		builders: make([]strings.Builder, n),
		prefix:   prefix,
		suffix:   suffix,
	}
	return buffer
}

func (tb *textBuffer) appendByIndex(index int, text string) {
	if index >= 0 && index < len(tb.builders) {
		tb.builders[index].WriteString(text)
	}
}

func (tb *textBuffer) String(separator string) string {
	var result strings.Builder

	for i, builder := range tb.builders {
		result.WriteString(tb.prefix)
		result.WriteString(builder.String())
		result.WriteString(tb.suffix)
		if i < len(tb.builders)-1 {
			result.WriteString(separator)
		}
	}

	return result.String()
}
