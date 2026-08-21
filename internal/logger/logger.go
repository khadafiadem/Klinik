package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

var (
	Info  *log.Logger
	Error *log.Logger
	Debug *log.Logger
)

func Init(level string) {
	flags := log.Ldate | log.Ltime

	Info = log.New(os.Stdout, "[INFO] ", flags)
	Error = log.New(os.Stderr, "[ERROR] ", flags)
	Debug = log.New(os.Stdout, "[DEBUG] ", flags)

	if strings.ToLower(level) != "debug" {
		Debug.SetOutput(io.Discard)
	}
}
