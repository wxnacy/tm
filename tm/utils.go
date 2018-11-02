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

func deleteFromString(s string, index, length int) string {
    if index + length - 1 >= len(s) || length <= 0{
        return s
    }
    return s[0:index] + s[index + length:]
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
