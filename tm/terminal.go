package tm

import (
    "github.com/nsf/termbox-go"
    "os"
    "github.com/mattn/go-runewidth"
    "strings"
    "fmt"
    "time"
)


type Mode uint8
type Position uint8

const (
    ModeInsert Mode = iota
    ModeNormal
    ModeCommand
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
    name string
    width, height    int
    lastCursorX, lastCursorY int
    cursorX, cursorY int
    tableSplitSymbolPosition int
    resultsSplitSymbolPosition int
    e *Event
    position Position
    mode Mode
    isListenKeyBorad bool

    tables []string
    tablesShowBegin int
    tablesLastCursorY int

    results [][]string
    resultsFormat []string
    resultsShowBegin int
    resultsBottomContent string
    resultTotalCount int
    resultsIsError bool
    resultsFormatIfNeedRefresh bool

    commands []string
    commandsSources []string
    commandsShowBegin int
    commandsBottomContent string
    commandsMode Mode
    commandsClipboard []string
    commandsWidth, commandsHeight int
    commandsLastCursorX, commandsLastCursorY int

    isShowFrame bool
    frame []string
    framePositionX, framePositionY int
    frameWidth, frameHeight int

    cells [][]Cell
    viewCells [][]Cell
    onOpenTable func(name string)
    onExecCommands func(cmds []string)
}

func New(name string) (*Terminal, error){
    err := termbox.Init()
    if err != nil {
        return nil, err
    }

    w, h := termbox.Size()

    t := &Terminal{
        name: name,
        width: w,
        height: h,
        tableSplitSymbolPosition: 20,
        resultsSplitSymbolPosition: 6,
        e: &Event{},
        mode: ModeNormal,
        position: PositionTables,
        cursorY: 1,
        isListenKeyBorad: true,

        tables: make([]string, 0),
        tablesShowBegin: 0,
        tablesLastCursorY: 1,

        results: make([][]string, 0),
        resultsFormat: make([]string, 0),
        resultsShowBegin: 0,
        resultsBottomContent: "",
        resultsIsError: false,
        resultsFormatIfNeedRefresh: false,

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

    cmd_file := cmdPath(this.name)
    if IsFile(cmd_file) {
        data, err := ReadFile(cmd_file)
        checkErr(err)
        this.commandsSources = strings.Split(data, "\n")
    }
}

func (this *Terminal) OnOpenTable(onOpenTable func(name string) ) {
    this.onOpenTable = onOpenTable
}

func (this *Terminal) OnExecCommands(onExecCommands func(cmds []string)) {
    this.onExecCommands = onExecCommands
}

func (this *Terminal) IsListenKeyBorad() bool {
    return this.isListenKeyBorad
}

func (this *Terminal) SetTables(tables []string) {
    this.tables = tables
}

func (this *Terminal) ClearResults() {
    r := make([][]string, 0)
    this.SetResults(r)

}

func (this *Terminal) SetResults(results [][]string) {
    this.results = results
    this.resultsFormatIfNeedRefresh = true
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
    if this.resultsFormatIfNeedRefresh {
        this.resultsFormat = mysqlArrayResultsFormat(this.results)
    }
    this.resultsFormatIfNeedRefresh = false
    b := this.resultsFormat

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

    _, resultsHeight := this.resultsSize()
    begin := time.Now()
    for y := 0; y < resultsHeight; y++ {

        index := y + this.resultsShowBegin + 2
        if index >= len(b) {
            Log.Info("reset result half y time: ", time.Since(begin))
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

    Log.Info("reset result time: ", time.Since(begin))

}

func (this *Terminal) initCommands() {
    this.commands = make([]string, 0)

    lineNumWidth := this.commandsLineNumWidth()

    prefix := fmt.Sprintf("%%%dd", lineNumWidth - 1)

    for i, d := range this.commandsSources {
        cmd := fmt.Sprintf(prefix + " %s", i + 1, d)
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
        if i > cy {
            return
        }
        if index >= len(this.commands) {
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

            if cellsX < len(this.cells[i]) {
                this.cells[i][cellsX] = Cell{
                    Ch: line[j],
                    Fg: fg,
                    Bg: bg,
                }
            }


        }

    }
}

func (this *Terminal) reset() {

    this.resetField()
    this.clearCells()
    this.resetTables()
    this.initCommands()
    this.resetCommands()
    this.resetResults()
    this.resetViewCells()
    this.resetCursor()

}

func (this *Terminal) resetField() {

    this.commandsWidth = this.width - this.tableSplitSymbolPosition - 1
    this.commandsHeight = this.resultsSplitSymbolPosition - 1
}

func (this *Terminal) resetCursor() {
    switch this.position {
        case PositionCommands: {
            if this.commandsMode == ModeCommand {
                return
            }
            maxCX, maxCY := this.commandsMaxCursor()
            minCX, _ := this.commandsMinCursor()
            if this.cursorX > maxCX {
                this.cursorX = maxCX
            }
            if this.cursorY > maxCY {
                this.cursorY = maxCY
            }
            if this.cursorX < minCX {
                this.cursorX = minCX
            }
        }
    }

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
    tb := time.Now()
    this.reset()
    begin := time.Now()
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
    Log.Infof("Rendering flush time: %v", time.Since(begin))
    Log.Infof("Rendering total time: %v", time.Since(tb))
}

func (this *Terminal) ListenKeyBorad() {

    e := this.PollEvent()
    switch e.Key {
        // case termbox.KeyEsc: {
            // if ! this.isCursorInCommands() {
                // os.Exit(0)
            // }
        // }
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
            if this.position == PositionResults {

                this.cursorY += 2
            }
        }
        case 'K': {
            if this.resultsSplitSymbolPosition <= 4 {
                return
            }
            this.resultsSplitSymbolPosition -= 2
            if this.position == PositionResults {

                this.cursorY -= 2
            }
        }
        case 'H': {
            if this.tableSplitSymbolPosition <= 3 {
                return
            }
            this.tableSplitSymbolPosition--
            if this.position == PositionTables {
                return
            }
            this.cursorX--
        }
        case 'L': {
            if this.tableSplitSymbolPosition >= this.width / 2 {
                return
            }
            this.tableSplitSymbolPosition++
            if this.position == PositionTables {
                return
            }
            this.cursorX++
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
            this.isListenKeyBorad = false
            this.ClearResults()
            this.SetResultsBottomContent("Waiting")
            this.Rendering()

            t := this.currentTable()
            cmds := []string{fmt.Sprintf("select * from %s", t)}
            this.onExecCommands(cmds)
            this.moveCursorToResults()
            this.isListenKeyBorad = true

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
            this.commandsSave()
            this.isListenKeyBorad = false
            this.ClearResults()
            this.SetResultsBottomContent("Waiting")
            this.Rendering()

            this.onExecCommands([]string{
                this.commandsSourceCurrentLine(),
            })
            this.resultsShowBegin = 0
            this.isListenKeyBorad = true
        }
        case termbox.KeyCtrlA: {
            cx, _ := this.commandsMinCursor()
            this.cursorX = cx
        }
        case termbox.KeyCtrlE: {
            cx, _ := this.commandsMaxCursor()
            this.cursorX = cx
        }
    }

    switch this.commandsMode {
        case ModeNormal: {
            this.listenCommandsNormal()
        }
        case ModeInsert: {
            this.listenCommandsInsert()
        }
        case ModeCommand: {
            switch e.Key {
                case termbox.KeyBackspace2: {
                    // cmd := this.commandsSources[this.cursorY]
                    x, _ := this.commandsCursor()
                    if x + this.commandsLineNumWidth() - 1 <= 0 {
                        return
                    }
                    // cmd = deleteFromString(cmd, x - 1, 1)
                    // this.commandsSources[this.cursorY] = cmd

                    this.commandsBottomContent = deleteFromString(
                        this.commandsBottomContent,
                        this.cursorX - this.tableSplitSymbolPosition - 2,
                        1,
                    )
                    this.cursorX--
                }
                case termbox.KeyEsc: {
                    this.commandsMode = ModeNormal
                    this.commandsBottomContent = ""
                }
                case termbox.KeyEnter: {
                    if this.commandsBottomContent == ":w" {
                        this.commandsSave()
                        this.commandsBottomContent = fmt.Sprintf(
                            "\"%s\" %dL written",
                            cmdPath(this.name), this.commandsLength(),
                        )
                    } else {
                        this.commandsBottomContent = ""
                    }
                    this.commandsMode = ModeNormal
                    this.cursorX = this.lastCursorX
                    this.cursorY = this.lastCursorY
                }
            }

            if e.Ch <= 0 {
                return
            }

            this.commandsBottomContent = insertInString(
                this.commandsBottomContent,
                this.cursorX - this.tableSplitSymbolPosition,
                string(e.Ch),
            )
            this.cursorX++

        }
    }
}
func (this *Terminal) listenCommandsInsert() {

    e := this.e.e
    switch e.Key {
        case termbox.KeyBackspace2: {
            this.commandsDeleteByBackspace()
        }
        case termbox.KeyCtrlW: {
            currentPosition := this.commandsSourceCurrentLinePosition()
            cmd := this.commandsSources[currentPosition]
            cx, _ := this.commandsCursor()

            newcmd := deleteStringByCtrlW(cmd, cx)
            this.commandsSources[currentPosition] = newcmd
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
            cmd := this.commandsSources[this.cursorY]

            newCmds := splitStringByIndex(cmd, cx)
            this.commandsSources[this.cursorY] = newCmds[0]
            this.commandsSources = insertInStringArray(
                this.commandsSources,
                this.commandsSourceCurrentLinePosition() + 1, newCmds[1],
            )
            if this.cursorY == this.commandsHeight - 1 {
                this.commandsShowBegin++
            }
            if this.cursorY < this.commandsHeight - 1 {
                this.cursorY++
            }
            this.cursorX = minCX
        }
    }

    if this.e.ch <= 0 {
        return
    }
    this.insertToCommands()
}
func (this *Terminal) listenCommandsNormal() {
    e := this.e.e

    // switch e.Key {
        // case termbox.KeyEsc: {
            // os.Exit(0)
        // }
    // }
    if e.Ch <= 0 {
        return
    }

    switch this.e.ch {
        case 'q': {
            os.Exit(0)
        }
        case 'd': {
            if this.e.preCh == 'd' {
                this.commandsDeleteCurrentLine()
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
                this.cursorY + 1 + this.commandsShowBegin,
                "",
            )
            if this.cursorY == this.commandsHeight - 1 {
                this.commandsShowBegin++
            } else {
                this.cursorY++
            }
            minCX, _ := this.commandsMinCursor()
            this.cursorX = minCX
        }
        case ':': {
            this.lastCursorX = this.cursorX
            this.lastCursorY = this.cursorY
            this.commandsMode = ModeCommand
            this.commandsBottomContent = ":"
            minCX, _ := this.commandsMinCursor()
            this.cursorX = minCX - 2
            this.cursorY = this.commandsHeight
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
                cmd := this.commandsSourceCurrentLine()
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
                this.commandsSourceCurrentLinePosition() + 1,
                this.commandsClipboard[0],
            )
            this.cursorY++
        }
        case 'w': {
            nowX, _ := this.commandsCursor()
            cx := stringNextWordBegin(
                this.commandsSourceCurrentLine(), nowX,
            )
            minCX, _ := this.commandsMinCursor()
            this.cursorX = cx + minCX
        }
        case 'e': {
            nowX, _ := this.commandsCursor()
            cx := stringNextWordEnd(
                this.commandsSourceCurrentLine(), nowX,
            )
            minCX, _ := this.commandsMinCursor()
            this.cursorX = cx + minCX
        }
        case 'b': {
            nowX, _ := this.commandsCursor()
            cx := stringPreWordBegin(
                this.commandsSourceCurrentLine(), nowX,
            )
            minCX, _ := this.commandsMinCursor()
            this.cursorX = cx + minCX
        }
    }
}

func (this *Terminal) insertToCommands() {
    currentLineNum := this.commandsSourceCurrentLinePosition()
    cmd := this.commandsSources[currentLineNum]
    x, _ := this.commandsCursor()
    cmd = insertInString(
        cmd, x, string(this.e.ch),
    )
    this.commandsSources[currentLineNum] = cmd
    this.cursorX +=1

}
func (this *Terminal) commandsDeleteByBackspace() {

    currentLineNum := this.commandsSourceCurrentLinePosition()
    cmd := this.commandsSources[currentLineNum]
    x, _ := this.commandsCursor()
    if x <= 0 {
        this.commandsDeleteCurrentLine()
        maxCX := this.commandsMaxCursorXByCursorY(this.cursorY - 1)
        this.cursorX = maxCX
        if this.cursorY == 2 && this.commandsShowBegin > 0 {
            this.commandsShowBegin--
        } else {
            this.cursorY--
        }

        return
    }
    cmd = deleteFromString(cmd, x - 1, 1)
    this.commandsSources[currentLineNum] = cmd
    this.cursorX--

}

func (this *Terminal) commandsDeleteCurrentLine() {

    minCX, _ := this.commandsMinCursor()
    if len(this.commandsSources) == 1 {
        this.commandsSources = []string{""}
        this.cursorX = minCX
        return
    }
    this.commandsSources = deleteFromStringArray(
        this.commandsSources,
        this.commandsSourceCurrentLinePosition(), 1,
    )

    Log.Info("dd ", this.cursorY, this.commandsShowBegin, len(this.commandsSources))
    if this.cursorY == len(this.commandsSources) - this.commandsShowBegin{
        if this.commandsShowBegin > 0 {
            this.commandsShowBegin--
        } else {
            this.cursorY--
        }
    }
    this.cursorX = minCX
}
func (this *Terminal) commandsSourceCurrentLine() string {
    return this.commandsSources[this.commandsSourceCurrentLinePosition()]
}

func (this *Terminal) commandsSourceCurrentLinePosition() int {
    return this.cursorY + this.commandsShowBegin
}

func (this *Terminal) commandsSave() {
    SaveFile(cmdPath(this.name), strings.Join(this.commandsSources, "\n"))
}
func (this *Terminal) commandsCursor() (x, y int) {
    minCX, _ := this.commandsMinCursor()
    return this.cursorX - minCX, this.cursorY
}

func (this *Terminal) commandsPosition() (x, y int) {
    return this.tableSplitSymbolPosition + 1, 0
}
func (this *Terminal) commandsLineNumWidth() (int) {

    numLength := 1
    cmdsLength := len(this.commandsSources)
    if cmdsLength >= 10 && cmdsLength < 100 {
        numLength = 2
    }
    if cmdsLength >= 100 && cmdsLength < 1000 {
        numLength = 3
    }

    return numLength + 1
}

func (this *Terminal) commandsMinCursor() (int, int) {
    return this.tableSplitSymbolPosition + 1 + this.commandsLineNumWidth(), 0
}
func (this *Terminal) commandsMaxCursorXByCursorY(y int) (int) {

    cx, _ := this.commandsMinCursor()
    var x int

    line := this.commandsSources[y + this.commandsShowBegin]
    if len(line) == 0 {
        x = cx
    } else {
        lineNumWidth := this.commandsLineNumWidth()
        x = cx + min(this.commandsWidth - lineNumWidth, len(line)) - 1
        if this.commandsMode == ModeInsert {
            x++
        }
    }
    return x
}
func (this *Terminal) commandsMaxCursor() (int, int) {
    return this.commandsMaxCursorXByCursorY(this.cursorY), this.commandsMaxCursorY()
}
func (this *Terminal) commandsMaxCursorY() (int) {
    var y int

    if len(this.commands) == 0 {
        y = 0
    } else {
        y = min(this.commandsHeight, len(this.commands)) - 1
    }
    return y
}

func (this *Terminal) commandsLength() (int) {
    return len(this.commandsSources)
}
func (this *Terminal) commandsMaxShowBegin() (int) {
    _, cy := this.commandsMaxCursor()
    if len(this.commands) < 1 + cy {
        return 0
    }
    return len(this.commands) - 1 - cy
}
func (this *Terminal) resultsSize() (int, int) {
    x := this.width - this.tableSplitSymbolPosition - 2
    y := this.height - this.resultsSplitSymbolPosition - 2
    return x, y
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
    if this.position == PositionCommands {
        this.commandsLastCursorX = this.cursorX
        this.commandsLastCursorY = this.cursorY
    }

    minCX, minCY := this.resultsMinCursor()
    this.cursorX = minCX
    this.cursorY = minCY
    this.position = PositionResults
}

func (this *Terminal) moveCursorToTables() {
    if this.position == PositionCommands {
        this.commandsLastCursorX = this.cursorX
        this.commandsLastCursorY = this.cursorY
    }
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
    this.cursorX = this.commandsLastCursorX
    this.cursorY = this.commandsLastCursorY
    this.position = PositionCommands
}

func (this *Terminal) isCursorInTables() bool {
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


    Log.Infof(
        "nowx %d nowy %d width %d height %d",
        nowX, nowY, this.width, this.height,
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
            maxCX, maxCY := this.commandsMaxCursor()
            if nowX < mi  {
                return
            }

            if nowY == 0 {
                if this.commandsShowBegin > 0 && offsetY < 0 {
                    this.commandsShowBegin += offsetY
                }
            }

            if nowY > maxCY {
                if this.commandsShowBegin < this.commandsMaxShowBegin() {
                    this.commandsShowBegin += offsetY
                }
                nowY = maxCY
            }

            if nowX > maxCX {
                nowX = maxCX
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
        e := termbox.PollEvent()
        Log.Infof(
            "e.Type %v e.Key %v e.Ch %v",
            e.Type, e.Key, e.Ch,
        )
        switch e.Type {
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



