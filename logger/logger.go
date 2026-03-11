package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

func timestamp() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s[INFO]%s  %s\n", gray+timestamp()+reset, green, reset, msg)
}

func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s[WARN]%s  %s\n", gray+timestamp()+reset, yellow, reset, msg)
}

func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s[ERROR]%s %s\n", gray+timestamp()+reset, red, reset, msg)
}

func Fatal(format string, args ...any) {
	Error(format, args...)
	os.Exit(1)
}

func Debug(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s[DEBUG]%s %s\n", gray+timestamp()+reset, cyan, reset, msg)
}
