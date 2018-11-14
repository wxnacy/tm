package tm

import (
    "os"
    "strings"
    "io/ioutil"
    "errors"
    "fmt"
    "log"
)

var TM_DIR = os.Getenv("HOME") + "/.tm"
var LOG_DIR = TM_DIR + "/logs"
var CMD_DIR = TM_DIR + "/commands"
var TABLE_DIR = TM_DIR + "/tables"

func cmdPath(name string) string {
    return fmt.Sprintf("%s/%s", CMD_DIR, name)
}


func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return Exists(path) && !IsDir(path)
}

func SaveFile(path, content string) error{
    paths := strings.Split(path, "/")
    dir := strings.Join(paths[0:len(paths) - 1], "/")
    if !IsDir(dir) {
        err := os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return err
        }
    }
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    f.WriteString(content)
    defer f.Close()

    return nil
}

func ReadFile(path string) ( string, error) {
    if IsFile(path) {
        d, err := ioutil.ReadFile(path)
        if err != nil {
            return "", err
        }
        return string(d), nil
    }
    return "", errors.New(path + "is not exists")
}

func LogFile(str ...string) {
    path := LOG_DIR + "/tm.log"
    if !IsDir(LOG_DIR) {
        err := os.MkdirAll(LOG_DIR, os.ModePerm)
        checkErr(err)
    }
    file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    file.WriteString(strings.Join(str, " ") + "\n")
}

func InitLogger() {

    path := LOG_DIR + "/tm.log"
    if !IsDir(LOG_DIR) {
        err := os.MkdirAll(LOG_DIR, os.ModePerm)
        checkErr(err)
    }
    file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    log.SetOutput(file)
}

