package tm

import (
    "os"
    "testing"
)


func TestSaveFile(t *testing.T) {
    dir, _ := os.Getwd()
    path := dir + "/test/test"
    SaveFile(path, "test")

    flag := IsFile(path)
    if ! flag {
        t.Error(path + "is not exists")
    }

    path = dir + "/test/test1"
    SaveFile(path, "test")

    flag = IsFile(path)
    if ! flag {
        t.Error(path + "is not exists")
    }

    os.RemoveAll(dir + "/test")
}
