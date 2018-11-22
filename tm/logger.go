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
    var fmtstr []string
    for _ = range v {
        fmtstr = append(fmtstr, "%v")
    }
    this.log("INFO", strings.Join(fmtstr, " "), v...)
}

func (this *Logger) Infof(fmts string, v ...interface{}) {
    this.log("INFO", fmts, v...)
}


func (this *Logger) Error(v ...interface{}) {
    var fmtstr []string
    for _ = range v {
        fmtstr = append(fmtstr, "%v")
    }
    this.log("ERROR", strings.Join(fmtstr, " "), v...)
}

func (this *Logger) Errorf(fmts string, v ...interface{}) {
    this.log("INFO", fmts, v...)
}

func (this *Logger) log(level, fmts string, v ...interface{}) {
    _, filename, line, ok := runtime.Caller(2)
    content := fmt.Sprintf(fmts, v...)
    if ok {
        filenames := strings.Split(filename, "/")
        filename = filenames[len(filenames)-1]
    }
    l.Printf("[%s:%d\t] [%s] %s", filename, line, level, content)
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

