package log

import (
	"fmt"
	"io"
	"path"
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
	Er(err ServerError)
}

type logger struct {
	writer io.Writer
	prefix string
}

type fatalLogger struct {
	io.Writer
}

func Init(errorHandle, warningHandle, infoHandle, traceHandle io.Writer) {
	Fatal = &fatalLogger{}
	Error = &logger{writer: errorHandle, prefix: "ERROR"}
	Warning = &logger{writer: warningHandle, prefix: "WARNING"}
	Info = &logger{writer: infoHandle, prefix: "INFO"}
	Trace = &logger{writer: traceHandle, prefix: "TRACE"}
}

func (lg *logger) Pf(format string, v ...interface{}) {
	filename, line := getCaller()

	fmt.Fprintf(lg.writer, "%v: %v %v:%v: %v\n", lg.prefix, getTime(), filename, line, fmt.Sprintf(format, v...))
}

func (lg *logger) Er(err ServerError) {
	e, desc := err.Error(), err.Desc()
	if e != nil && desc != nil {
		fmt.Fprintf(lg.writer, "%v: %v %v:%v: error: %v, desc: %v\n", lg.prefix, getTime(), *err.File(), err.Line(), e, *desc)
	} else if e != nil {
		fmt.Fprintf(lg.writer, "%v: %v %v:%v: %v\n", lg.prefix, getTime(), *err.File(), err.Line(), e)
	} else if desc != nil {
		fmt.Fprintf(lg.writer, "%v: %v %v:%v: %v\n", lg.prefix, getTime(), *err.File(), err.Line(), *desc)
	}
}

func (lg *fatalLogger) Pf(format string, v ...interface{}) {
	filename, line := getCaller()

	panic(fmt.Sprintf("%v %v:%v: %v", getTime(), filename, line, fmt.Sprintf(format, v...)))
}

func (lg *fatalLogger) Er(err ServerError) {
	e, desc := err.Error(), err.Desc()
	if e != nil && desc != nil {
		panic(fmt.Sprintf("%v %v:%v: error: %v, desc: %v\n", getTime(), *err.File(), err.Line(), e, *desc))
	} else if e != nil {
		panic(fmt.Sprintf("%v %v:%v: %v\n", getTime(), *err.File(), err.Line(), e))
	} else if desc != nil {
		panic(fmt.Sprintf("%v %v:%v: %v\n", getTime(), *err.File(), err.Line(), *desc))
	}
}

func getTime() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05.99999")
}

func getCaller() (string, int) {
	_, p, line, ok := runtime.Caller(2)
	if ok {
		_, filename := path.Split(p)
		return filename, line
	}

	return "unknown file", -1
}
