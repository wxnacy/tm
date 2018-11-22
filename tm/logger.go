package tm

import (
    "log"
    "os"
    "runtime"
    "strings"
    "fmt"
)

var l *log.Logger
var Log = NewLogger()

type Logger struct {

}

func NewLogger() *Logger {
    L := &Logger{}

    if l == nil {
        l = initLogger()
    }

    return L
}

func (this *Logger) Info(v ...interface{}) {
    this.Infof("%p", v)
}

func (this *Logger) Infof(fmts string, v ...interface{}) {
    _, filename, line, _ := runtime.Caller(2)
    filenames := strings.Split(filename, "/")
    s := fmt.Sprintf(fmts, v...)
    l.Printf("[%s:%d\t] [INFO] %s", filenames[len(filenames)-1], line, s)
}

func (this *Logger) Error(v ...interface{}) {
    this.Errorf("%p", v)
}

func (this *Logger) Errorf(fmts string, v ...interface{}) {
    _, filename, line, _ := runtime.Caller(2)
    filenames := strings.Split(filename, "/")
    s := fmt.Sprintf(fmts, v...)
    l.Printf("[%s:%d\t] [ERROR] %s", filenames[len(filenames)-1], line, s)
}

func initLogger() *log.Logger{

    path := LOG_DIR + "/tm.log"
    if !IsDir(LOG_DIR) {
        os.MkdirAll(LOG_DIR, os.ModePerm)
        // err := os.MkdirAll(LOG_DIR, os.ModePerm)
        // checkErr(err)
    }
    file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    return log.New(file, "", log.LstdFlags)
}

