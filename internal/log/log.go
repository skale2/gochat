package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
)

var infoLog, warnLog, errorLog, fatalLog *log.Logger

func FileWriter() io.Writer {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return file
}

func Initialize(out io.Writer) {
	infoLog = log.New(out, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLog = log.New(out, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(out, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	fatalLog = log.New(out, "[FATAL] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(v ...interface{}) {
	infoLog.Output(2, fmt.Sprint(v...))
}

func Infof(s string, v ...interface{}) {
	infoLog.Output(2, fmt.Sprintf(s, v...))
}

func Warn(v ...interface{}) {
	warnLog.Output(2, fmt.Sprintln(v...))
}

func Warnf(s string, v ...interface{}) {
	warnLog.Output(2, fmt.Sprintf(s, v...))
}

func Error(v ...interface{}) {
	errorLog.Output(2, fmt.Sprintln(v...)+string(debug.Stack()))
}

func Errorf(s string, v ...interface{}) {
	errorLog.Output(2, fmt.Sprintf(s, v...)+string(debug.Stack()))
}

func Fatal(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	fatalLog.Output(2, msg)
	panic(msg)
}

func Fatalf(s string, v ...interface{}) {
	msg := fmt.Sprintf(s, v...)
	fatalLog.Output(2, msg)
	panic(msg)
}
