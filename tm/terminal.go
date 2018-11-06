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
    preKey termbox.Key
    key termbox.Key
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
    commandsSources []string
    commandsShowBegin int
    commandsBottomContent string
    commandsMode Mode
    commandsClipboard []string

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
        resultsSplitSymbolPosition: 6,
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

        commandsSources: []string{
            "select * from ad",
            "select * from config",
        },
        commands: make([]string, 0),
        commandsShowBegin: 0,
        commandsBottomContent: "",
        commandsMode: ModeNormal,
        commandsClipboard: make([]string, 0),

        cells: make([][]Cell, 0),
        viewCells: make([][]Cell, 0),
    }
    t.initArgs()

    return t, nil
}

func (this *Terminal) initArgs() {
    this.resultsSplitSymbolPosition = initResultsSplitSymbolPosition(this.height)
    LogFile(
        "result", strconv.Itoa(this.resultsSplitSymbolPosition),
    )
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
    maxTableLength := this.tableSplitSymbolPosition - 1
    titleEnd := len(titleCells)
    if len(titleCells) > maxTableLength {
        titleEnd = maxTableLength
    }
    this.cells[0] = cellsReplace(
        this.cells[0], 0,
        titleCells[0:titleEnd],
    )

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

            rbg := termbox.ColorDefault
            rfg := termbox.ColorDefault

            yu := 1
            if this.height % 2 == 1 {
                yu = 0
            }
            if oy % 4 == yu {
                rbg = termbox.ColorBlack
            }

            if this.isCursorInResults() {

                if oy == this.cursorY {
                    rbg = termbox.ColorYellow
                }
            }

            if oy + 1 < this.height && ox + 1 < this.width{
                this.cells[oy][ox] = Cell{
                    Ch: chs[x],
                    Fg: rfg,
                    Bg: rbg,
                }
            }
        }
    }

}

func (this *Terminal) initCommands() {
    this.commands = make([]string, 0)
    for i, d := range this.commandsSources {
        cmd := fmt.Sprintf("%d %s", i + 1, d)
        this.commands = append(this.commands, cmd)
    }
}

func (this *Terminal) resetCommands() {
    px, _ := this.resultsPosition()
    cy := this.commandsMaxCursorY()
    minCX, _ := this.commandsMinCursor()

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
        // LogFile("enter", strconv.Itoa(index), strconv.Itoa(this.commandsShowBegin))
        if i > cy {
            return
        }
        line := []rune(this.commands[index])
        for j := 0; j < len(line); j++ {
            cellsX := this.tableSplitSymbolPosition + j + 1
            fg := termbox.ColorDefault
            bg := termbox.ColorDefault
            if cellsX < minCX  {
                bg = termbox.ColorBlack
            }

            if this.isCursorInCommands() && i == this.cursorY{

                if cellsX >= minCX{
                    bg = termbox.ColorBlack
                } else {
                    bg = termbox.ColorDefault
                }
            }

            this.cells[i][cellsX] = Cell{
                Ch: line[j],
                Fg: fg,
                Bg: bg,
            }
        }

    }
}

func (this *Terminal) reset() {
    this.clearCells()
    this.resetTables()
    this.initCommands()
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
}

func (this *Terminal) ListenKeyBorad() {

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
        case 'g': {
            if this.e.preCh == 'g' {
                this.moveCursorToFirstLine()
            }
        }
        case 'G': {
            this.moveCursorToLastLine()
        }
        case 'J': {
            if this.resultsSplitSymbolPosition >= this.height - 6{
                return
            }
            this.resultsSplitSymbolPosition += 2
        }
        case 'K': {
            if this.resultsSplitSymbolPosition <= 2 {
                return
            }
            this.resultsSplitSymbolPosition -= 2
        }
        case 'H': {
            if this.tableSplitSymbolPosition <= 3 {
                return
            }
            this.tableSplitSymbolPosition--
        }
        case 'L': {
            if this.tableSplitSymbolPosition >= this.width / 2 {
                return
            }
            this.tableSplitSymbolPosition++
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
        case 'o': {
            t := this.currentTable()
            cmds := []string{fmt.Sprintf("select * from %s limit 20", t)}
            this.onExecCommands(cmds)
            this.moveCursorToResults()
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
                    minCX, _ := this.commandsMinCursor()
                    if len(this.commandsSources) == 1 {
                        this.commandsSources = []string{""}
                        this.cursorX = minCX
                        return
                    }
                    this.commandsSources = deleteFromStringArray(
                        this.commandsSources,
                        this.cursorY, 1,
                    )

                    cy := min(len(this.commandsSources) - 1, this.cursorY)
                    this.cursorY = cy
                    this.cursorX = minCX
                    this.e.ch = 0
                }
            }
            case 'x': {

                cmd := this.commandsSources[this.cursorY]
                x, _ := this.commandsCursor()
                if x < 0 {
                    return
                }
                cmd = deleteFromString(cmd, x, 1)
                LogFile(cmd)
                this.commandsSources[this.cursorY] = cmd
            }
            case 'i': {
                this.commandsMode = ModeInsert
                this.commandsBottomContent = "-- INSERT --"
            }
            case 'o': {
                this.commandsMode = ModeInsert
                this.commandsBottomContent = "-- INSERT --"

                this.commandsSources = insertInStringArray(
                    this.commandsSources,
                    this.cursorY + 1,
                    "",
                )
                this.cursorY++
                minCX, _ := this.commandsMinCursor()
                this.cursorX = minCX
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
            case 'y': {
                if this.e.preCh == 'y' {
                    cmd := this.commandsSources[this.cursorY]
                    this.commandsClipboard = append(
                        []string{cmd}, this.commandsClipboard...,
                    )
                }
            }
            case 'p': {
                if len(this.commandsClipboard) == 0 {
                    return
                }
                this.commandsSources = insertInStringArray(
                    this.commandsSources,
                    this.cursorY + 1,
                    this.commandsClipboard[0],
                )
                this.cursorY++
            }
            case 'w': {
                nowX, _ := this.commandsCursor()
                cx := stringNextWordBegin(
                    this.commandsSources[this.cursorY], nowX,
                )
                minCX, _ := this.commandsMinCursor()
                this.cursorX = cx + minCX
            }
            case 'e': {
                nowX, _ := this.commandsCursor()
                cx := stringNextWordEnd(
                    this.commandsSources[this.cursorY], nowX,
                )
                minCX, _ := this.commandsMinCursor()
                this.cursorX = cx + minCX
            }
            case 'b': {
                nowX, _ := this.commandsCursor()
                cx := stringPreWordBegin(
                    this.commandsSources[this.cursorY], nowX,
                )
                minCX, _ := this.commandsMinCursor()
                this.cursorX = cx + minCX
            }
        }

    } else {

        switch e.Key {
            case termbox.KeyBackspace2: {
                cmd := this.commandsSources[this.cursorY]
                x, _ := this.commandsCursor()
                if x <= 0 {
                    return
                }
                cmd = deleteFromString(cmd, x - 1, 1)
                LogFile(cmd)
                this.commandsSources[this.cursorY] = cmd
                this.cursorX--
            }
            case termbox.KeyCtrlW: {
                cmd := this.commandsSources[this.cursorY]
                cx, _ := this.commandsCursor()

                newcmd := deleteStringByCtrlW(cmd, cx)
                this.commandsSources[this.cursorY] = newcmd
                if len(cmd) > len(newcmd) {
                    this.cursorX -= len(cmd) - len(newcmd)
                }
            }
            case termbox.KeyEsc: {
                this.commandsMode = ModeNormal
                this.commandsBottomContent = ""
            }
            case termbox.KeyEnter: {
                cx, _ := this.commandsCursor()
                minCX, _ := this.commandsMinCursor()
                _, maxCY := this.commandsMaxCursor()
                cmd := this.commandsSources[this.cursorY]

                newCmds := splitStringByIndex(cmd, cx)
                LogFile(newCmds[0], newCmds[1])
                this.commandsSources[this.cursorY] = newCmds[0]
                this.commandsSources = insertInStringArray(
                    this.commandsSources,
                    this.cursorY + 1, newCmds[1],
                )
                LogFile(
                    "keyenter",
                    strconv.Itoa(maxCY),
                )
                if this.cursorY == this.resultsSplitSymbolPosition - 2 {
                    this.commandsShowBegin++
                }
                if this.cursorY < this.resultsSplitSymbolPosition - 2{
                    this.cursorY++
                }
                this.cursorX = minCX
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
    cmd := this.commandsSources[this.cursorY]
    LogFile("before", cmd)
    x, _ := this.commandsCursor()
    cmd = insertInString(
        cmd, x, string(this.e.ch),
    )
    LogFile(cmd)
    this.commandsSources[this.cursorY] = cmd
    this.cursorX +=1

}

func (this *Terminal) commandsCursor() (x, y int) {
    minCX, _ := this.commandsMinCursor()
    return this.cursorX - minCX, this.cursorY
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

func (this *Terminal) resultsMinCursor() (int, int) {
    // results 区域最小的光标坐标位置
    var x, y int
    x = this.tableSplitSymbolPosition + 1
    y = this.resultsSplitSymbolPosition + 3
    return x, y
}

func (this *Terminal) resultsMaxCursor() (int, int) {
    // results 区域最大的光标坐标位置
    // var x, y int
    // x = this.tableSplitSymbolPosition + 1
    // y = this.resultsSplitSymbolPosition + 3
    y := min(
        this.height - 3,
        this.resultsSplitSymbolPosition + len(this.results) * 2 - 1,
    )
    return this.width - 1, y
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
    minCX, minCY := this.resultsMinCursor()
    this.cursorX = minCX
    this.cursorY = minCY
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

            _, maxCY := this.resultsMaxCursor()

            if this.height >= len(this.results) + 1 + this.resultsSplitSymbolPosition + 1 {
                this.cursorY = maxCY
            } else {

                this.cursorY = maxCY
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
            _, py := this.resultsPosition()
            _, maxCY := this.resultsMaxCursor()

            if nowY < py + 2 {
                if offsetY < 0 && this.resultsShowBegin > 0 {
                    this.resultsShowBegin += offsetY
                }
                return
            } else if nowY > maxCY {
                if this.resultsShowBegin < len(this.results) + this.resultsSplitSymbolPosition + 3 - this.height {

                    this.resultsShowBegin += offsetY
                }
                nowY = maxCY
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

                t.e.preKey = t.e.key
                t.e.key = e.Key
                return e
            case termbox.EventResize:
                t.width = e.Width
                t.height = e.Height
                t.tablesShowBegin = 0
                t.tablesLastCursorY = 1
                t.initArgs()
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
