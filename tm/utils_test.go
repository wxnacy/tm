package tm

import (
    "testing"
    "fmt"
    "strings"
)

func TestCellsToString(t *testing.T) {
    cells := []Cell{
        Cell{Ch: 'h'},
        Cell{Ch: 'e'},
        Cell{Ch: 'l'},
        Cell{Ch: 'l'},
        Cell{Ch: 'o'},
    }

    res := cellsToString(cells)
    if res != "hello" {
        t.Error(res + " is Error")
    }
}

func TestInsertInString(t *testing.T) {

    s := "select"
    res := ""
    res = insertInString(s, 0, string(rune('a')))
    if res != "aselect"{
        t.Error(res + "is error")
    }

    res = insertInString(s, 1, "aa")
    if res != "saaelect"{
        t.Error(res + "is error")
    }

    res = insertInString(s, 6, "aa")
    if res != "selectaa"{
        t.Error(res + "is error")
    }

    res = insertInString("select ", 7, "aa")
    if res != "select aa"{
        t.Error(res + "is error")
    }

}

func TestDeleteFromString(t *testing.T) {

    s := "select"
    res := ""
    res = deleteFromString(s, 0, 1)
    if res != "elect"{
        t.Error(res + "is error")
    }

    res = deleteFromString(s, 1, 2)
    if res != "sect"{
        t.Error(res + "is error")
    }

    res = deleteFromString(s, 5, 2)
    if res != "select"{
        t.Error(res + "is error")
    }

    res = deleteFromString(s, 2, -1)
    if res != "select"{
        t.Error(res + "is error")
    }

    res = deleteFromString(s, 7, 1)
    if res != "select"{
        t.Error(res + "is error")
    }

}
func TestDeleteStringByCtrlW(t *testing.T) {

    s := "select * from  user"
    res := ""
    res = deleteStringByCtrlW(s, 3)
    if res != "ect * from  user"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 6)
    if res != " * from  user"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 7)
    if res != "* from  user"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 16)
    if res != "select * fromser"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 23)
    if res != "select * from  user"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 0)
    if res != "select * from  user"{
        t.Error(res + "is error")
    }
    res = deleteStringByCtrlW(s, -1)
    if res != "select * from  user"{
        t.Error(res + "is error")
    }
}

func TestInsertInStringArray(t *testing.T) {
    var arr = []string{"1", "2"}
    var newArr = make([]string, 0)

    newArr = insertInStringArray(arr, 0, "0")
    fmt.Println(newArr)
    if strings.Join(newArr, "") != "012" {
        t.Error(newArr, "is error")
    }


    newArr = insertInStringArray(arr, 1, "4")
    fmt.Println(newArr)
    if strings.Join(newArr, "") != "142" {
        t.Error(newArr, "is error")
    }

    newArr = insertInStringArray(arr, 3, "3")
    fmt.Println(newArr)
    if strings.Join(newArr, "") != "123" {
        t.Error(newArr, "is error")
    }
}

func TestinitResultsSplitSymbolPosition(t *testing.T) {
    var p int
    p = initResultsSplitSymbolPosition(18)
    if p != 6 {
        t.Error(p, "is error")
    }

    p = initResultsSplitSymbolPosition(55)
    if p != 19 {
        t.Error(p, "is error")
    }
    p = initResultsSplitSymbolPosition(59)
    if p != 19 {
        t.Error(p, "is error")
    }
}
