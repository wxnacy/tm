package tm

import (
    "strings"
    "github.com/nsf/termbox-go"
)

func cellsToString(cells []Cell) string {
    chs := make([]rune, 0)
    for _, d := range cells {
        chs = append(chs, d.Ch)
    }
    return string(chs)
}

func stringToCells(s string) []Cell {
    return stringToCellsWithColor(s, termbox.ColorDefault, termbox.ColorDefault)
}

func stringToCellsWithColor(s string, fg termbox.Attribute, bg termbox.Attribute) []Cell {
    chs := []rune(s)
    cells := make([]Cell, 0)
    for _, d := range chs {
        cells = append(cells, Cell{Ch: d, Fg: fg, Bg: bg})
    }
    return cells
}

func cellsReplace(cells []Cell, index int, newCells []Cell) []Cell{

    for i, d := range newCells {
        x := i + index
        if x >= len(cells) {
            return cells
        }
        cells[i + index] = d
    }
    return cells

}

func insertInString(s string, x int, apd string) string {
    if x >= len(s) {
        return s + apd
    } else {
        return s[0:x] + apd + s[x:]
    }
}

func insertInStringArray(arr []string, index int, s string) []string {
    // 在字符串数组中添加字符串
    if index == 0 {
        return append([]string{s}, arr...)
    } else if index >= len(arr) {
        return append(arr, s)
    } else {
        return append(arr[0:index], append([]string{s}, arr[index:]...)...)
    }
}

func splitStringByIndex(s string, index int) []string {
    if index == 0 {
        return []string{"", s}
    } else if index >= len(s) {
        return []string{s, ""}
    } else {
        return []string{s[0:index], s[index:]}
    }
}

func deleteFromString(s string, index, length int) string {
    if index + length - 1 >= len(s) || length <= 0{
        return s
    }
    return s[0:index] + s[index + length:]
}

func deleteFromStringArray(arr []string, index, length int) []string {
    if index == 0 {
        return arr[index + length:]
    } else if index >= len(arr) {
        return arr
    } else {
        return append(arr[0:index] , arr[index + length:]...)
    }
}

func stringNextWordBegin(s string, index int) int {
    if index >= len(s) {
        return index
    }
    splits := strings.Split(s[index:], " ")
    if len(splits) <= 1 {
        return index
    }
    return strings.Index(s, splits[1])
}

func stringNextWordEnd(s string, index int) int {
    if index >= len(s) {
        return index
    }
    splits := strings.Split(s[index:], " ")
    if len(splits) == 0 {
        return index
    }

    end := splits[0]
    if len(end) <= 1 {
        if len(splits) > 1 {
            end = splits[1]
        } else {
            return index
        }
    } else {
        return index + len(end) - 1
    }

    if len(end) < 1 {
        return index
    }

    return strings.Index(s[index:], end) + len(end) + index - 1
}

func stringPreWordBegin(s string, index int) int {
    if index == 0 || index >= len(s) {
        return index
    }

    splits := strings.Split(s[0:index], " ")

    if len(splits) == 1 {
        return 0
    }

    indexStr := splits[len(splits) - 1]
    if indexStr == "" {
        indexStr = splits[len(splits) - 2]
    }

    return strings.LastIndex(s[0:index], indexStr)
}

func deleteStringByCtrlW(s string, index int) string {
    if index > len(s) || index <= 0{
        return s
    }

    prefix := s[0:index]

    prefixs := strings.Split(prefix, " ")
    // fmt.Println(prefixs)

    if len(prefixs) == 1 {
        return s[index:]
    }

    begin := len(prefixs) - 2
    for i := len(prefixs) - 2; i >= 0; i-- {
        if prefixs[i] != ""{
            begin = i
            break
        }
    }

    prefix_index := begin + 1
    if prefixs[len(prefixs) - 1] == "" {
        prefix_index = begin
    }

    return strings.Join(prefixs[0:prefix_index], " ") + s[index:]
}


func mysqlArrayResultsFormat(a [][]string) []string {

    widths := make([]int, 0)
    for i := 0; i < len(a); i++ {
        line := a[i]
        width := len(line)
        for j := 0; j < width; j++ {
            if len(widths) < width {
                widths = append(widths, len(line[j]))
            } else {
                if widths[j] < len(line[j]) {
                    max := len(line[j])
                    if max > 18 {
                        max = 18
                    }
                    widths[j] = max
                }
            }
        }
    }

    b := make([]string, 0)
    for i := 0; i < len(a); i++ {
        line := a[i]
        newLine := ""
        for j := 0; j < len(line); j++ {
            suffixLen := widths[j] - len(line[j])
            suffix := ""
            if suffixLen > 0 {
                suffix = strings.Repeat(" ", suffixLen)
            }

            item := line[j]
            if widths[j] < len(item) {
                end := widths[j]
                if end >= len(item) {
                    end = len(item) - 1
                }
                item = item[0:end]
            }
            newLine +=" " + item + suffix + " |"
        }
        b = append(b, newLine)
        b = append(b, strings.Repeat("-", len(newLine)))
    }
    return b
}

func initResultsSplitSymbolPosition(height int) int {

    yu := height % 2

    res := height / 3
    resYu := res % 2

    if yu == resYu {
        return res
    } else {
        return res + 1
    }

}

func min(x, y int) int {
    if x <= y {
        return x
    }
    return y
}

func max(x, y int) int {
    if x >= y {
        return x
    }
    return y
}
