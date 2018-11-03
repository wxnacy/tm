package tm

import (
    "github.com/nsf/termbox-go"
    "os"
    "github.com/mattn/go-runewidth"
    "strings"
    "strconv"
    "fmt"
)

type Mode uint8
type Position uint8

const (
    ModeInsert Mode = iota
    ModeNormal
    PositionTables Position = iota
    PositionCommands
    PositionResults
)

type Event struct {
    preCh rune
    ch rune
    e termbox.Event
}

type Cell struct {
    Ch rune
    Fg termbox.Attribute    // 文字颜色
    Bg termbox.Attribute    // 背景颜色
}

func (c *Cell) Width() int {
    return runewidth.RuneWidth(c.Ch)
}

type Terminal struct {
    width, height    int
    cursorX, cursorY int
    tableSplitSymbolPosition int
    resultsSplitSymbolPosition int
    e *Event
    position Position
    mode Mode

    tables []string
    tablesShowBegin int
    tablesLastCursorY int

    results [][]string
    resultsShowBegin int
    resultsBottomContent string
    resultTotalCount int
    resultsIsError bool

    commands []string
    commandsShowBegin int
    commandsBottomContent string
    commandsMode Mode

    cells [][]Cell
    viewCells [][]Cell
    onOpenTable func(name string)
    onExecCommands func(cmds []string)
}

func New() (*Terminal, error){
    err := termbox.Init()
    if err != nil {
        return nil, err
    }

    w, h := termbox.Size()

    t := &Terminal{
        width: w,
        height: h,
        tableSplitSymbolPosition: 20,
        resultsSplitSymbolPosition: 5,
        e: &Event{},
        mode: ModeNormal,
        position: PositionTables,
        cursorY: 1,
        tables: make([]string, 0),
        tablesShowBegin: 0,
        tablesLastCursorY: 1,
        results: make([][]string, 0),
        resultsShowBegin: 0,
        resultsBottomContent: "",
        resultsIsError: false,

        commands: []string{
            "1 select * from ad",
            "2 select * from shop",
            "3 select * from config;",
            // "4 select * from config;",
            // "5 select * from config;",
            // "6 select * from config;",
        },
        commandsShowBegin: 0,
        commandsBottomContent: "",
        commandsMode: ModeNormal,

        cells: make([][]Cell, 0),
        viewCells: make([][]Cell, 0),
    }

    return t, nil
}

func (this *Terminal) OnOpenTable(onOpenTable func(name string) ) {
    this.onOpenTable = onOpenTable
}

func (this *Terminal) OnExecCommands(onExecCommands func(cmds []string)) {
    this.onExecCommands = onExecCommands
}

func (this *Terminal) SetTables(tables []string) {
    this.tables = tables
}

func (this *Terminal) SetResults(results [][]string) {
    this.results = results
}

func (this *Terminal) SetResultsIsError(flag bool) {
    this.resultsIsError = flag
}

func (this *Terminal) SetResultsBottomContent(content string) {
    this.resultsBottomContent = content
}

func (this *Terminal) resetTables() {
    // tables := append([]string{"Tables:"}, this.tables...)

    tables := this.tables
    if len(tables) == 0 {
        return
    }

    LogFile(strconv.Itoa(this.tablesShowBegin))
    titleCells := stringToCells(tables[0])
    this.cells[0] = cellsReplace(this.cells[0], 0, titleCells)

    for y := 1; y < len(tables); y++ {
        prefix := ""
        if y > 0 {
            prefix = strings.Repeat(" ", 1)
        }

        index := y + this.tablesShowBegin
        if index >= len(tables) {
            return
        }
        chs := []rune(prefix + tables[y + this.tablesShowBegin] + strings.Repeat(" ", this.tableSplitSymbolPosition))
        for x := 0; x < len(chs); x++ {
            if y + 1 < this.height && x +1 < this.tableSplitSymbolPosition {
                bg := termbox.ColorDefault
                if y == this.tablesLastCursorY {
                    bg = termbox.ColorYellow
                }
                this.cells[y][x] = Cell{Ch: chs[x], Bg: bg}
            }
        }
    }
}

func (this *Terminal) resetResults() {
    // reset bottom
    fg := termbox.ColorCyan
    bg := termbox.ColorDefault
    if this.resultsIsError {
        fg = termbox.ColorRed
    }

    this.cells[this.height - 1] = cellsReplace(
        this.cells[this.height - 1],
        this.tableSplitSymbolPosition + 2,
        stringToCellsWithColor(this.resultsBottomContent, fg, bg),
    )

    if len(this.results) == 0 {
        return
    }
    b := mysqlArrayResultsFormat(this.results)

    this.cells[this.resultsSplitSymbolPosition + 1] = cellsReplace(
        this.cells[this.resultsSplitSymbolPosition + 1],
        this.tableSplitSymbolPosition + 1,
        stringToCells(b[0]),
    )
    this.cells[this.resultsSplitSymbolPosition + 2] = cellsReplace(
        this.cells[this.resultsSplitSymbolPosition + 2],
        this.tableSplitSymbolPosition + 1,
        stringToCells(b[1]),
    )

    for y := 0; y < len(b); y++ {

        index := y + this.resultsShowBegin + 2
        if index >= len(b) {
            return
        }
        chs := []rune(b[index])
        for x := 0; x < len(chs); x++ {
            oy := this.resultsSplitSymbolPosition + y + 3
            ox := this.tableSplitSymbolPosition + 1 + x

            if oy + 1 < this.height && ox + 1 < this.width{
                this.cells[oy][ox] = Cell{Ch: chs[x]}
            }
        }
    }

}

func (this *Terminal) resetCommands() {
    px, _ := this.resultsPosition()
    cy := this.commandsMaxCursorY()

    this.cells[this.resultsSplitSymbolPosition - 1] = cellsReplace(
        this.cells[this.resultsSplitSymbolPosition - 1],
        px,
        stringToCellsWithColor(
            this.commandsBottomContent,
            termbox.ColorBlue,
            termbox.ColorDefault,
        ),
    )

    for i := 0; i < len(this.commands); i++ {
        index := i + this.commandsShowBegin
        if i > cy {
            return
        }
        line := []rune(this.commands[index])
        for j := 0; j < len(line); j++ {
            this.cells[i][this.tableSplitSymbolPosition + j + 1] = Cell{Ch: line[j]}
        }

    }
}

func (this *Terminal) reset() {
    this.clearCells()
    this.resetTables()
    this.resetCommands()
    this.resetResults()
    this.resetViewCells()
}

func (this *Terminal) clearCells() {

    res := make([][]Cell, 0)

    for i := 0; i < this.height; i++ {
        viewLine := make([]Cell, 0)
        for j := 0; j < this.width; j++ {
            if i == this.resultsSplitSymbolPosition &&
                j > this.tableSplitSymbolPosition {
                viewLine = append(viewLine, Cell{Ch: '='})
            } else {
                viewLine = append(viewLine, Cell{Ch: ' '})
            }
        }
        viewLine[this.tableSplitSymbolPosition] = Cell{Ch: '|', Bg: termbox.ColorWhite}
        // viewLine[this.tableSplitSymbolPosition + 1] = Cell{Ch: '|'}
        res = append(res, viewLine)
    }

    this.cells = res
}


func (this *Terminal) stringToCellsWithColor(s string, fg, bg termbox.Attribute) [][]Cell {
    cells := make([][]Cell, 0)
    lines := strings.Split(s, "\n")
    if len(lines) < this.height {
        lines = append(lines, )
    }
    for _, d := range lines {
        ycells := this.stringToLineWithColor(d, fg, bg)
        cells = append(cells, ycells)
    }
    return cells
}

func (this *Terminal) stringToLineWithColor(s string, fg, bg termbox.Attribute) []Cell {
    cells := make([]Cell, 0)
    chs := []rune(s)

    for i := 0; i < this.width; i++ {
        var c rune
        if i + 1 < len(chs) {
            c = chs[i]
        } else {
            c = '1'
        }
        cell := Cell{
            Ch: c,
            Fg: fg,
            Bg: bg,
        }
        cells = append(cells, cell)
    }
    return cells
}

func (this *Terminal) AppendCellFromString(s string) {
    cells := this.stringToCellsWithColor(s, termbox.ColorDefault, termbox.ColorDefault)
    for _, d := range cells {
        this.cells = append(this.cells, d)
    }
    this.resetViewCells()
}

func (this *Terminal) Rendering() {
    this.reset()
    termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)

    for y, yd := range this.cells {
        x := 0
        LoopX:
        for xi := 0; xi < this.width; xi++ {
            if xi >= len(yd) {
                break LoopX
            }
            d := yd[xi]
            termbox.SetCell(x, y, d.Ch, d.Fg, d.Bg)
            x += d.Width()
        }
    }
    termbox.SetCursor(this.cursorX, this.cursorY)
    termbox.Flush()
    this.listenKeyBorad()
}

func (this *Terminal) listenKeyBorad() {

    e := this.PollEvent()
    switch e.Key {
        case termbox.KeyEsc: {
            if ! this.isCursorInCommands() {
                os.Exit(0)
            }
        }
        case termbox.KeyCtrlH: {
            if this.isCursorInResults() || this.isCursorInCommands(){
                this.moveCursorToTables()
            }
        }
        case termbox.KeyCtrlL: {
            if this.isCursorInTables() {
                this.moveCursorToCommands()
            }
        }
        case termbox.KeyCtrlJ: {
            if this.isCursorInCommands() {
                this.moveCursorToResults()
            }
        }
        case termbox.KeyCtrlK: {
            if this.isCursorInResults() {
                this.moveCursorToCommands()
            }
        }
        case termbox.KeyEnter: {
            if this.isCursorInTables() {
                this.onOpenTable(this.currentTable())
            }
        }
    }
    switch this.position {
        case PositionCommands: {
            this.listenCommands()
        }
        case PositionResults: {
            this.listenResults()
        }
        case PositionTables: {
            this.listenTables()
        }
    }

    if e.Ch <= 0 {
        return
    }

    switch this.mode {
        case ModeNormal: {
            this.listenModeNormal(e)
        }
    }

}

func (this *Terminal) listenModeNormal(e termbox.Event) {
    switch e.Ch {
        case 'q':{
            if ! this.isCursorInCommands() {
                os.Exit(0)
            }
        }
        case 'o': {
            if this.isCursorInTables() {
                t := this.currentTable()
                cmds := []string{fmt.Sprintf("select * from %s limit 20", t)}
                this.onExecCommands(cmds)
                this.moveCursorToResults()
            }
        }
        case 'g': {
            if this.e.preCh == 'g' {
                this.moveCursorToFirstLine()
            }
        }
        case 'G': {
            this.moveCursorToLastLine()
        }
    }

}

func (this *Terminal) listenTables() {
    e := this.e.e

    if e.Ch <= 0 {
        return
    }

    switch e.Ch {
        case 'j':{
            this.moveCursor(0, 1)
        }
        case 'k': {
            this.moveCursor(0, -1)
        }
    }

}

func (this *Terminal) listenResults() {
    e := this.e.e

    if e.Ch <= 0 {
        return
    }

    switch e.Ch {
        case 'j':{
            this.moveCursor(0, 2)
        }
        case 'k': {
            this.moveCursor(0, -2)
        }
    }

}

func (this *Terminal) listenCommands() {
    e := this.e.e

    switch e.Key {
        case termbox.KeyArrowLeft: {
            this.moveCursor(-1, 0)
        }
        case termbox.KeyArrowRight: {
            this.moveCursor(1, 0)
        }
        case termbox.KeyCtrlR: {
            cmd := this.commands[this.cursorY]
            LogFile(cmd[2:])

            this.onExecCommands([]string{cmd[2:]})
            this.resultsShowBegin = 0
        }
        case termbox.KeyCtrlA: {
            cx, _ := this.commandsMinCursor()
            this.cursorX = cx
        }
    }
    if this.commandsMode == ModeNormal {
        switch e.Key {
            case termbox.KeyEsc: {
                os.Exit(0)
            }
            case termbox.KeyCtrlE: {
                cx, _ := this.commandsMaxCursor()
                this.cursorX = cx
            }
        }
        if e.Ch <= 0 {
            return
        }

        switch this.e.ch {
            case 'q': {
                os.Exit(0)
            }
            case 'd': {
                if this.e.preCh == 'd' {
                    this.commands[this.cursorY] = fmt.Sprintf("%d ", this.cursorY + 1)
                    px, _ := this.resultsPosition()
                    this.cursorX = px + 2
                }
            }
            case 'x': {

                cmd := this.commands[this.cursorY]
                x, _ := this.commandsCursor()
                if x < 2 {
                    return
                }
                cmd = deleteFromString(cmd, x, 1)
                LogFile(cmd)
                this.commands[this.cursorY] = cmd
            }
            case 'i': {
                if this.isCursorInCommands() {
                    this.commandsMode = ModeInsert
                    this.commandsBottomContent = "-- INSERT --"
                }
            }
            case 'h': {
                this.moveCursor(-1, 0)
            }
            case 'l': {
                this.moveCursor(1, 0)
            }
            case 'j': {
                this.moveCursor(0, 1)
            }
            case 'k': {
                this.moveCursor(0, -1)
            }
            case 'g': {
                if this.e.preCh == 'g' {
                    this.cursorY = 0
                    this.commandsShowBegin = 0
                }
            }
            case 'G': {
                _, cy := this.commandsMaxCursor()
                this.cursorY = cy
                this.commandsShowBegin = this.commandsMaxShowBegin()
            }
        }

    } else {

        switch e.Key {
            case termbox.KeyBackspace2: {
                cmd := this.commands[this.cursorY]
                x, _ := this.commandsCursor()
                if x <= 2 {
                    return
                }
                cmd = deleteFromString(cmd, x - 1, 1)
                LogFile(cmd)
                this.commands[this.cursorY] = cmd
                this.cursorX--
            }
            case termbox.KeyCtrlW: {
                cmd := this.commands[this.cursorY]
                cmdstr := cmd[2:]

                newcmd := deleteStringByCtrlW(cmdstr, this.cursorX - this.tableSplitSymbolPosition - 3)
                this.commands[this.cursorY] = cmd[0:2] + newcmd
                if len(cmdstr) > len(newcmd) {
                    this.cursorX -= len(cmdstr) - len(newcmd)
                }
            }
            case termbox.KeyEsc: {
                this.commandsMode = ModeNormal
                this.commandsBottomContent = ""
            }
            case termbox.KeyEnter: {
                LogFile("end")
                px, _ := this.resultsPosition()
                cmd := this.commands[this.cursorY]
                this.commands[this.cursorY] = cmd[0:this.cursorX - px]
                newLine := fmt.Sprintf(
                    "%d %s", len(this.commands) + 1, 
                    cmd[this.cursorX - px:],
                )
                this.commands = append(this.commands, newLine)
                this.cursorY++
                this.cursorX = px + 2
            }
            case termbox.KeyCtrlE: {
                cx, _ := this.commandsMaxCursor()
                this.cursorX = cx + 1
            }
        }

        if this.e.ch <= 0 {
            return
        }
        this.insertToCommands()
    }
}

func (this *Terminal) insertToCommands() {
    cmd := this.commands[this.cursorY]
    LogFile("before", cmd)
    x, _ := this.commandsCursor()
    cmd = insertInString(
        cmd, x, string(this.e.ch),
    )
    LogFile(cmd)
    this.commands[this.cursorY] = cmd
    this.cursorX +=1

}

func (this *Terminal) commandsCursor() (x, y int) {
    return this.cursorX - this.tableSplitSymbolPosition - 1, this.cursorY
}

func (this *Terminal) commandsPosition() (x, y int) {
    return this.tableSplitSymbolPosition + 1, 0
}
func (this *Terminal) commandsMinCursor() (int, int) {
    return this.tableSplitSymbolPosition + 3, 0
}
func (this *Terminal) commandsMaxCursor() (int, int) {
    cx, _ := this.commandsMinCursor()
    var x int

    line := this.commands[this.cursorY]
    if len(line) == 2 {
        x = cx
    } else {
        x = cx + min(this.width - this.tableSplitSymbolPosition, len(line)) - 3
    }
    return x, this.commandsMaxCursorY()
}
func (this *Terminal) commandsMaxCursorY() (int) {
    var y int

    if len(this.commands) == 0 {
        y = 0
    } else {
        y = min(this.resultsSplitSymbolPosition - 1, len(this.commands)) - 1
    }
    return y
}
func (this *Terminal) commandsMaxShowBegin() (int) {
    _, cy := this.commandsMaxCursor()
    if len(this.commands) < 1 + cy {
        return 0
    }
    return len(this.commands) - 1 - cy
}
func (this *Terminal) resultsPosition() (x, y int) {
    return this.tableSplitSymbolPosition + 1, this.resultsSplitSymbolPosition + 1
}

func (this *Terminal) moveCursorToFirstLine() {
    switch this.position {
        case PositionTables: {
            this.cursorY = 1
            this.tablesShowBegin = 0
            this.tablesLastCursorY = this.cursorY
        }
        case PositionResults: {
            _, py := this.resultsPosition()
            this.cursorY = py + 2
            this.resultsShowBegin = 0
        }
    }
}

func (this *Terminal) moveCursorToResults() {
    switch this.position {
        case PositionTables: {
            this.tablesLastCursorY = this.cursorY
            this.resultsShowBegin = 0
        }
    }
    px, py := this.resultsPosition()
    this.cursorX = px
    this.cursorY = py + 2
    this.position = PositionResults
}

func (this *Terminal) moveCursorToTables() {
    this.cursorX = 0
    this.cursorY = this.tablesLastCursorY
    this.position = PositionTables
}

func (this *Terminal) moveCursorToCommands() {
    switch this.position {
        case PositionTables: {
            this.tablesLastCursorY = this.cursorY
        }
    }
    cx, cy := this.commandsMinCursor()
    this.cursorX = cx
    this.cursorY = cy
    this.position = PositionCommands
}

func (this *Terminal) isCursorInTables() bool {
    LogFile(strconv.Itoa(this.cursorX), strconv.Itoa(this.cursorY))
    return this.cursorX < this.tableSplitSymbolPosition && this.cursorY > 0
}

func (this *Terminal) isCursorInResults() bool {
    return this.cursorX > this.tableSplitSymbolPosition && this.cursorY > this.resultsSplitSymbolPosition
}

func (this *Terminal) isCursorInCommands() bool {
    return this.cursorX > this.tableSplitSymbolPosition && this.cursorY < this.resultsSplitSymbolPosition
}
func (this *Terminal) currentTable() string{
    currentTable := this.tables[this.cursorY + this.tablesShowBegin]
    LogFile("table name ", currentTable)
    return currentTable
}

func (this *Terminal) moveCursorToLastLine() {
    switch this.position {
        case PositionTables: {
            if this.height - 1 >= len(this.tables) {
                this.cursorY = len(this.tables) - 1
            } else {
                this.cursorY = this.height - 2
                this.tablesShowBegin = len(this.tables) - this.height
            }
            this.tablesLastCursorY = this.cursorY
        }
        case PositionResults: {

            if this.height >= len(this.results) + 1 + this.resultsSplitSymbolPosition + 1 {
                _, py := this.resultsPosition()
                this.cursorY = py + len(this.results) * 2 - 2
            } else {

                this.cursorY = this.height - 3
                this.resultsShowBegin = len(this.results) * 2 - (this.height - this.resultsSplitSymbolPosition)
            }
        }
    }
}

func (this *Terminal) moveCursor(offsetX, offsetY int) {

    nowX := this.cursorX
    nowX += offsetX
    if nowX < 0 {
        return
    }

    nowY := this.cursorY
    nowY += offsetY
    if nowY < 0 {
        nowY = 0
    }

    LogFile(
        "nowx", strconv.Itoa(nowX), "nowy", strconv.Itoa(nowY),
        "width", strconv.Itoa(this.width), "height", strconv.Itoa(this.height),
    )


    switch this.position {
        case PositionTables: {
            if nowY >= len(this.tables) {
                return
            }
            if nowY > this.height - 2 {
                if this.tablesShowBegin >= len(this.tables) - this.height + 1{
                    return
                }
                this.tablesShowBegin += nowY - this.height + 2
                nowY = this.height - 2
            } else if nowY == 0 {
                if offsetY < 0  && this.tablesShowBegin > 0{
                    this.tablesShowBegin += offsetY
                }
                return
            }

            this.tablesLastCursorY = nowY

        }
        case PositionCommands: {
            px, _ := this.commandsPosition()
            mi := px
            _, cy := this.commandsMaxCursor()
            if nowX < mi  {
                return
            }

            if nowY == 0 {
                if this.commandsShowBegin > 0 && offsetY < 0 {
                    this.commandsShowBegin += offsetY
                }
            }

            if nowY > cy {
                if this.commandsShowBegin < this.commandsMaxShowBegin() {
                    this.commandsShowBegin += offsetY
                }
                nowY = cy
            }


        }
        case PositionResults: {
            // TODO 移动还有问题
            _, py := this.resultsPosition()

            if nowY < py + 2 {
                if offsetY < 0 && this.resultsShowBegin > 0 {
                    this.resultsShowBegin += offsetY
                }
                return
            } else if nowY > this.height - 3 {
                if this.resultsShowBegin < len(this.results) + this.resultsSplitSymbolPosition + 3 - this.height {

                    this.resultsShowBegin += offsetY
                }
                nowY = this.height - 2
            }
        }
    }

    this.cursorX = nowX
    this.cursorY = nowY


}

func (t *Terminal) PollEvent() termbox.Event{
    for {
        switch e := termbox.PollEvent(); e.Type {
            case termbox.EventKey:
                t.e.preCh = t.e.ch
                if e.Key == termbox.KeySpace {
                    t.e.ch = ' '
                } else {
                    t.e.ch = e.Ch
                }
                t.e.e = e
                return e
            case termbox.EventResize:
                t.width = e.Width
                t.height = e.Height
                t.tablesShowBegin = 0
                t.tablesLastCursorY = 1
                t.moveCursorToTables()
                t.Rendering()
        }
    }
}

func (this *Terminal) resetViewCells() {


    for i := 0; i < len(this.cells); i++ {
        viewLine := make([]Cell, 0)
        line := this.cells[i]
        for j := 0; j < len(line); j++ {
            viewLine = append(viewLine, line[j])
        }
        this.viewCells = append(this.viewCells, viewLine)
    }

}

func (this *Terminal) Close() {
    termbox.Close()
}
func LogFile(str ...string) {
    file, _ := os.OpenFile("wsh.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    file.WriteString(strings.Join(str, " ") + "\n")
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
