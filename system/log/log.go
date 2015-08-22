package log

import (
	"fmt"
	"io"
	"runtime"
	"time"
)

var (
	Fatal   Logger
	Error   Logger
	Warning Logger
	Info    Logger
	Trace   Logger
)

type Logger interface {
	Pf(format string, v ...interface{})
}

type logger struct {
	writer io.Writer
	prefix string
}

type fatalLogger struct {
	io.Writer
}

func Init(fatalHandle, errorHandle, warningHandle, infoHandle, traceHandle io.Writer) {
	Fatal = &fatalLogger{fatalHandle}
	Error = &logger{writer: errorHandle, prefix: "ERROR"}
	Warning = &logger{writer: warningHandle, prefix: "WARNING"}
	Info = &logger{writer: infoHandle, prefix: "INFO"}
	Trace = &logger{writer: traceHandle, prefix: "TRACE"}
}

func (lg *logger) Pf(format string, v ...interface{}) {
	filename, line := getCaller()

	fmt.Fprintf(lg.writer, "%v: %v %v:%v: %v\n", lg.prefix, getTime(), filename, line, fmt.Sprintf(format, v...))
}

func (lg *fatalLogger) Pf(format string, v ...interface{}) {
	filename, line := getCaller()

	fmt.Fprintf(lg, "FATAL: %v %v:%v: %v\n", getTime(), filename, line, fmt.Sprintf(format, v...))
	panic(fmt.Sprintf("%v %v:%v: %v", getTime(), filename, line, fmt.Sprintf(format, v...)))
}

func getTime() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05.99999")
}

func getCaller() (string, int) {
	_, filename, line, ok := runtime.Caller(2)
	if ok {
		return filename, line
	}

	return "unknown file", -1
}
