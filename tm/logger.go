package tm

import (
    "log"
    "os"
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
    l.SetPrefix("[INFO] ")
    l.Print(v...)
}

func (this *Logger) Infof(fmt string, v ...interface{}) {
    l.SetPrefix("[INFO] ")
    l.Printf(fmt, v...)
}

func (this *Logger) Error(v ...interface{}) {
    l.SetPrefix("[ERROR] ")
    l.Print(v...)
}

func (this *Logger) Errorf(fmt string, v ...interface{}) {
    l.SetPrefix("[ERROR] ")
    l.Printf(fmt, v...)
}

func initLogger() *log.Logger{

    path := LOG_DIR + "/tm.log"
    if !IsDir(LOG_DIR) {
        os.MkdirAll(LOG_DIR, os.ModePerm)
        // err := os.MkdirAll(LOG_DIR, os.ModePerm)
        // checkErr(err)
    }
    file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    return log.New(file, "", log.LstdFlags|log.Lshortfile)
}

