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

// func TestCommandToCells(t *testing.T) {
    // s := " hahah ssss sss   dd  "
    // if  commandToCells(s, termbox.ColorDefault) == nil {
        // t.Error("error")
    // }
// }

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

    if res != "select * from  ser"{
        t.Error(res + "is error")
    }

    res = deleteStringByCtrlW(s, 23)
    if res != "select * from  "{
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

func TestDeleteFromStringArray(t *testing.T) {
    var arr1 = []string{
        "0",
        "1",
        "2",
        "3",
    }

    var newArr []string

    newArr = deleteFromStringArray(arr1, 0, 1)
    if strings.Join(newArr, "") != "123" {
        t.Error(newArr, "is error")
    }

    newArr = deleteFromStringArray(arr1, 1, 2)
    if strings.Join(newArr, "") != "03" {
        t.Error(newArr, "is error")
    }
    arr1 = []string{
        "0",
        "1",
        "2",
        "3",
    }
    newArr = deleteFromStringArray(arr1, 4, 2)
    if strings.Join(newArr, "") != "0123" {
        t.Error(newArr, "is error")
    }

}

func TestStringNextWordBegin(t *testing.T) {
    var s string
    var i int
    s = "select * from user"
    i = stringNextWordBegin(s, 0)
    if i != 7 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordBegin(s, 3)
    if i != 7 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordBegin(s, 7)
    if i != 9 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordBegin(s, 15)
    if i != 15 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordBegin(s, 25)
    if i != 25 {
        t.Error(i, "is error")
    }
}

func TestStringNextWordEnd(t *testing.T) {
    var s string
    var i int
    s = "select * from user"
    i = stringNextWordEnd(s, 0)
    if i != 5 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 3)
    if i != 5 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 7)
    if i != 12 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 8)
    if i != 12 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 11)
    if i != 12 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 15)
    if i != 17 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringNextWordEnd(s, 25)
    if i != 25 {
        t.Error(i, "is error")
    }
}
func TestStringPreWordBegin(t *testing.T) {
    var s string
    var i int
    s = "select * from user"
    i = stringPreWordBegin(s, 0)
    if i != 0 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 3)
    if i != 0 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 7)
    if i != 0 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 8)
    if i != 7 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 6)
    if i != 0 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 10)
    if i != 9 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 15)
    if i != 14 {
        t.Error(i, "is error")
    }
    s = "from * from user"
    i = stringPreWordBegin(s, 13)
    if i != 12 {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWordBegin(s, 18)
    if i != 14 {
        t.Error(i, "is error")
    }
    s = "select * from user  "
    i = stringPreWordBegin(s, 20)
    if i != 14 {
        t.Error(i, "is error")
    }
    s = "select   * from user"
    i = stringPreWordBegin(s, 9)
    if i != 0 {
        t.Error(i, "is error")
    }
}

func TestStringPreWord(t *testing.T) {
    var s string
    var i string
    s = "select * from user"
    i = stringPreWord(s, 0)
    if i != "" {
        t.Error(i, "is error")
    }

    s = "select * from user"
    i = stringPreWord(s, 2)
    if i != "se" {
        t.Error(i, "is error")
    }

    s = "select * from user"
    i = stringPreWord(s, 7)
    if i != "select" {
        t.Error(i, "is error")
    }

    s = "select * from user"
    i = stringPreWord(s, 7)
    if i != "select" {
        t.Error(i, "is error")
    }
    s = "select * from user"
    i = stringPreWord(s, 9)
    if i != "*" {
        t.Error(i, "is error")
    }

    s = "select   * from user"
    i = stringPreWord(s, 9)
    if i != "select" {
        t.Error(i, "is error")
    }

    s = "select user  "
    i = stringPreWord(s, 12)
    if i != "user" {
        t.Error(i, "is error")
    }
    s = "select user  "
    i = stringPreWord(s, 14)
    if i != "user" {
        t.Error(i, "is error")
    }
}

func TestInArray(t *testing.T) {
    var i int
    i = inArray(1, []int{3, 2, 1})
    if i != 2 {
        t.Error(i, " is error")
    }
    i = inArray(4, []int{3, 2, 1})
    if i != -1 {
        t.Error(i, " is error")
    }
    i = inArray(3, []int{3, 2, 1})
    if i != 0 {
        t.Error(i, " is error")
    }
    i = inArray("1", []string{"3", "1"})
    if i != 1 {
        t.Error(i, " is error")
    }
}
