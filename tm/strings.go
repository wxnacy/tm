package tm

import (
    "strings"
)


func FilterStrings(array []string, s string) (results []string) {
    results = make([]string, 0)
    if s == "" {
        results = array
        return
    }


    var newArr []string
    var tempArr []string
    // var offset = 0


    var zeroCount int
    begin := 0
    var end int
    newArr = array
    for i := 0; i < len(s); i++ {
        end = i + 1 + begin
        if end > len(s) {
            return
        }
        tempArr = filterStringItem(newArr, s[begin:end], begin)
        Log.Info(results)
        if len(tempArr) == 0 {
            // newArr = filterStringItem(newArr, s[begin:end-1], begin)
            newArr = results
            begin = i
            i = 0
            zeroCount++
        } else {
            results = tempArr
        }
        Log.Info(zeroCount)
        if zeroCount > len(s) {
            Log.Info("result")
            return
        }
    }

    return
}

func filterStringItem(array []string, s string, begin int) (results []string) {
    results = make([]string, 0)

    for _, d := range array {
        slice := d[begin:]
        if strings.HasPrefix(slice, s) && inArray(d, results) == -1{
            results = append(results, d)
        }
    }
    for _, d := range array {
        slice := d[begin:]
        if strings.HasPrefix(strings.ToLower(slice), strings.ToLower(s)) && inArray(d, results) == -1{
            results = append(results, d)
        }
    }

    for _, d := range array {
        slice := d[begin:]
        if strings.Contains(slice, s) && inArray(d, results) == -1{
            results = append(results, d)
        }
    }

    for _, d := range array {
        slice := d[begin:]
        if strings.Contains(strings.ToLower(slice), strings.ToLower(s)) && inArray(d, results) == -1{
            results = append(results, d)
        }
    }
    return
}
