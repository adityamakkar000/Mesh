package ui

import (
	"fmt"
	"io"
	"strings"
)

const (
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Bold   = "\033[1m"
)

type PrefixWriter struct {
	prefix string
	out    io.Writer
}

func NewPrefixWriter(prefix string, out io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefix: prefix,
		out:    out,
	}
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			break
		}
		if _, err := fmt.Fprintf(w.out, "%s%s%s\n", w.prefix, line, Reset); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func Success(msg string) {
	fmt.Printf("%s==> %s%s\n", Green, msg, Reset)
}

func Error(msg string) {
	fmt.Printf("%sERROR: %s%s\n", Red, msg, Reset)
}

// Error(msg) wrapper that also returns fmt.Errorf with format "wrapped: err"
func ErrorWrap(err error, format string, args ...any) error {
	wrapped := fmt.Errorf(format, args...)
	Error(wrapped.Error())
	return fmt.Errorf("%w: %w", wrapped, err)
}

func Warn(msg string) {
	fmt.Printf("%sWARN: %s%s\n", Yellow, msg, Reset)
}

func Info(msg string) {
	fmt.Printf("%s==> %s%s\n", Cyan, msg, Reset)
}

func Header(msg string) {
	fmt.Printf("%s%s==> %s%s\n", Bold, Cyan, msg, Reset)
}
