package tm

import (
    "github.com/nsf/termbox-go"
    "os"
    "github.com/mattn/go-runewidth"
    "strings"
    "fmt"
    "time"
    "encoding/json"
)


type Mode uint8
type Position uint8
type FramesMode uint8

const (
    ModeInsert Mode = iota
    ModeNormal
    ModeCommand
    PositionTables Position = iota
    PositionCommands
    PositionResults

    FramesModeTables FramesMode = iota
    FramesModeResultsDetail
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
    resultsColumns []string
    resultsFormat []string
    resultsShowBegin int
    resultsLeftShowBegin int
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

    isShowFrames bool
    frames []string
    framesMode FramesMode
    framesPositionX, framesPositionY int
    framesWidth, framesHeight int
    framesHighlightLinePosition int
    framesShowBegin int

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
        resultsLeftShowBegin: 0,
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

        isShowFrames: false,
        frames: make([]string, 0),
        framesHighlightLinePosition: -1,
        framesShowBegin: 0,
        framesPositionX: 0,
        framesPositionY: 0,

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


func (this *Terminal) resetFrames() {
    if ! this.isShowFrames {
        return
    }

    // this.framesPositionX = this.cursorX + 2

    if this.cursorY < this.height / 2 {

        this.framesPositionY = this.cursorY + 1
    } else {
        this.framesPositionY = this.cursorY - this.framesHeight
    }

    for y := 0; y < this.framesHeight; y++ {

        bg := termbox.ColorWhite
        cy := this.framesPositionY + y
        framesIndex := y + this.framesShowBegin
        if framesIndex < len(this.frames) {

            chs := []rune(this.frames[framesIndex])

            if y == this.framesHighlightLinePosition {
                Log.Infof("y %d hith %d", y, this.framesHighlightLinePosition)
                bg = termbox.ColorGreen
            }

            for x := -1; x < this.framesWidth; x++ {
                cx := this.framesPositionX + 1 + x
                ch := ' '
                if x + 1 <= len(chs) && x > -1 {
                    ch = chs[x]
                }
                this.cells[cy][cx] = Cell{
                    Ch: ch,
                    Bg: bg,
                }
            }

        } else {

            for x := -1; x < this.framesWidth; x++ {
                cx := this.framesPositionX + 1 + x
                ch := ' '
                // if x + 1 <= len(chs) && x > -1 {
                    // ch = chs[x]
                // }
                this.cells[cy][cx] = Cell{
                    Ch: ch,
                    Bg: bg,
                }
            }
        }

    }

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
        this.resultsColumns = this.results[0]
        this.resultsFormat = mysqlArrayResultsFormat(this.results)
    }
    Log.Info("colums ", this.resultsColumns)
    this.resultsFormatIfNeedRefresh = false
    b := this.resultsFormat

    this.cells[this.resultsSplitSymbolPosition + 1] = cellsReplace(
        this.cells[this.resultsSplitSymbolPosition + 1],
        this.tableSplitSymbolPosition + 1,
        stringToCells(b[0][this.resultsLeftShowBegin:]),
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

            chsIndex := x + this.resultsLeftShowBegin
            if oy + 1 < this.height && ox + 1 < this.width &&
            chsIndex < len(chs){

                this.cells[oy][ox] = Cell{
                    Ch: chs[chsIndex],
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
    // minCX, _ := this.commandsMinCursor()

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

        bg := termbox.ColorDefault
        if this.isCursorInCommands() && i == this.cursorY{
            bg = termbox.ColorBlack
        }
        this.cells[i] = cellsReplace(
            this.cells[i],
            this.tableSplitSymbolPosition + 1,
            commandToCells(this.commands[index], bg),
        )
        // line := []rune(this.commands[index])
        // for j := 0; j < len(line); j++ {
            // cellsX := this.tableSplitSymbolPosition + j + 1
            // fg := termbox.ColorDefault
            // bg := termbox.ColorDefault
            // if cellsX < minCX  {
                // bg = termbox.ColorBlack
            // }

            // if this.isCursorInCommands() && i == this.cursorY{

                // if cellsX >= minCX{
                    // bg = termbox.ColorBlack
                // } else {
                    // bg = termbox.ColorDefault
                // }
            // }

            // if cellsX < len(this.cells[i]) {
                // this.cells[i][cellsX] = Cell{
                    // Ch: line[j],
                    // Fg: fg,
                    // Bg: bg,
                // }
            // }


        // }

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
    this.resetFrames()

}

func (this *Terminal) resetField() {

    this.commandsWidth = this.width - this.tableSplitSymbolPosition - 1
    this.commandsHeight = this.resultsSplitSymbolPosition - 1

    this.framesHeight = min(this.height / 2 - 1, len(this.frames))
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
        case termbox.KeyEsc: {
            this.isShowFrames = false
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
        case termbox.KeyTab: {
            if this.isShowFrames {
                this.framesMoveDown()
                if this.isCursorInCommands() {
                    word := this.frames[this.framesHighlightLinePosition]
                    Log.Infof("high %s", word)
                    if this.commandsPreWord() != "from" &&
                    this.commandsPreWord() != "update"{
                        this.commandsDeleteByCtrlW()
                    }
                    cx, _ := this.commandsCursor()
                    word = "`" + word + "`"
                    this.commandsInsert(cx, word)
                }

            }
        }
        case termbox.KeyArrowUp: {

            if this.isShowFrames {
                this.framesMoveUp()
            }
        }
        case termbox.KeyArrowDown: {
            if this.isShowFrames {
                this.framesMoveDown()
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
            this.resultsLeftShowBegin = 0
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
            if this.isShowFrames {
                this.framesInitForResultsDetail()
            }
        }
        case 'k': {
            this.moveCursor(0, -2)
            if this.isShowFrames {
                this.framesInitForResultsDetail()
            }
        }
        case 'l': {
            this.resultsMoveRight()
        }
        case 'h': {
            this.resultsMoveLeft()
        }
        case 'o': {
            this.framesInitForResultsDetail()
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
            this.resultsLeftShowBegin = 0
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

                    switch this.commandsBottomContent {
                        case ":w": {
                            this.commandsSave()
                            this.commandsBottomContent = fmt.Sprintf(
                                "\"%s\" %dL written",
                                cmdPath(this.name), this.commandsLength(),
                            )
                        }
                        case ":wq": {
                            this.commandsSave()
                            os.Exit(0)
                        }
                        case ":q": {
                            os.Exit(0)
                        }
                        default: {
                            this.commandsBottomContent = ""
                        }
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
            this.framesChangeByBackspace()
        }
        case termbox.KeyCtrlW: {
            this.commandsDeleteByCtrlW()
            this.framesChangeByBackspace()
        }
        case termbox.KeyEsc: {
            this.commandsMode = ModeNormal
            this.commandsBottomContent = ""
            this.isShowFrames = false
        }
        case termbox.KeyEnter: {
            cx, _ := this.commandsCursor()
            minCX, _ := this.commandsMinCursor()
            currentLineY := this.commandsSourceCurrentLinePosition()
            cmd := this.commandsSources[currentLineY]

            newCmds := splitStringByIndex(cmd, cx)
            this.commandsSources[currentLineY] = newCmds[0]
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
            this.isShowFrames = false
        }
    }

    if this.e.ch <= 0 {
        return
    }

    this.commandsInsertByKeyBorad()
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
            currentLineNum := this.commandsSourceCurrentLinePosition()
            cmd := this.commandsSources[currentLineNum]
            x, _ := this.commandsCursor()
            if x < 0 {
                return
            }
            cmd = deleteFromString(cmd, x, 1)
            this.commandsSources[currentLineNum] = cmd
        }
        case 'i': {
            this.commandsChangeMode(ModeInsert)
        }
        case 'I': {
            this.commandsChangeMode(ModeInsert)
            minCX, _ := this.commandsMinCursor()
            this.cursorX = minCX
        }
        case 'a': {
            this.commandsChangeMode(ModeInsert)
            this.cursorX++
        }
        case 'A': {
            this.commandsChangeMode(ModeInsert)
            maxCX, _ := this.commandsMaxCursor()
            this.cursorX = maxCX
        }
        case 'S': {
            this.commandsSources[this.commandsSourceCurrentLinePosition()] = ""
            this.commandsChangeMode(ModeInsert)
            minCX, _ := this.commandsMinCursor()
            this.cursorX = minCX
        }
        case 'c': {
            if this.e.preCh == 'c' {

                this.commandsSources[this.commandsSourceCurrentLinePosition()] = ""
                this.commandsChangeMode(ModeInsert)
                minCX, _ := this.commandsMinCursor()
                this.cursorX = minCX
                this.e.ch = 0
            }
        }
        case 'o': {
            this.commandsChangeMode(ModeInsert)

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
        case '0': {
            lineNumWidth := this.commandsLineNumWidth()
            this.cursorX = this.tableSplitSymbolPosition + 1 + lineNumWidth
        }
        case '$': {
            maxCX, _ := this.commandsMaxCursor()
            this.cursorX = maxCX
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

func (this *Terminal) commandsInsertByKeyBorad() {
    x, _ := this.commandsCursor()
    this.commandsInsert(x, string(this.e.ch))
    this.framesChangeByInsert()

}

func (this *Terminal) commandsInsert(index int, s string) (cmd string){
    currentLineNum := this.commandsSourceCurrentLinePosition()
    cmd = this.commandsSources[currentLineNum]
    cmd = insertInString(
        cmd, index, s,
    )
    this.commandsSources[currentLineNum] = cmd
    this.cursorX += len(s)
    return
}

func (this *Terminal) commandsDeletePreWord() (newcmd string){

    currentPosition := this.commandsSourceCurrentLinePosition()
    cmd := this.commandsSources[currentPosition]
    cx, _ := this.commandsCursor()
    newcmd = deleteStringByCtrlW(cmd, cx)
    this.commandsSources[currentPosition] = newcmd
    return
}


func (this *Terminal) commandsPreRune() (r rune){
    cx, _ := this.commandsCursor()
    r = 0
    if cx == 0 {
        return
    }

    cmd := this.commandsSourceCurrentLine()
    r = []rune(cmd)[cx- 1]
    return

}
func (this *Terminal) commandsPreWord() (word string){
    cx, _ := this.commandsCursor()
    word = stringPreWord(
        this.commandsSourceCurrentLine(),
        cx,
    )
    return
}

func (this *Terminal) commandsDeleteByCtrlW() (newcmd string){

    cmd := this.commandsSourceCurrentLine()
    newcmd = this.commandsDeletePreWord()
    if len(cmd) > len(newcmd) {
        this.cursorX -= len(cmd) - len(newcmd)
    }
    return
}

func (this *Terminal) commandsDeleteByBackspace() {

    currentLineNum := this.commandsSourceCurrentLinePosition()
    cmd := this.commandsSources[currentLineNum]
    x, y := this.commandsCursor()
    // Log.Infof("x %d Y %d minCX %d, minCY %d", x, y, minCX, minCY)
    if x == 0 && y == 0 {
        return
    }

    if x <= 0 {
        if cmd == "" {
            return
        }
        this.commandsDeleteCurrentString()
        maxCY := this.commandsMaxCursorY()
        if this.cursorY == maxCY && this.commandsShowBegin > 0 {
            this.commandsShowBegin--
        } else {
            this.cursorY--
        }

        maxCX, _ := this.commandsMaxCursor()
        this.cursorX = maxCX

        cx , _ := this.commandsCursor()
        this.commandsInsert(cx, cmd)
        this.cursorX -= len(cmd)
        return
    }
    cmd = deleteFromString(cmd, x - 1, 1)
    this.commandsSources[currentLineNum] = cmd
    this.cursorX--
}

func (this *Terminal) commandsDeleteCurrentString() (line string){

    line = this.commandsSourceCurrentLine()

    if len(this.commandsSources) == 1 {
        this.commandsSources = []string{""}
        return
    }
    this.commandsSources = deleteFromStringArray(
        this.commandsSources,
        this.commandsSourceCurrentLinePosition(), 1,
    )
    return
}
func (this *Terminal) commandsDeleteCurrentLine() (line string){
    line = this.commandsDeleteCurrentString()

    minCX, _ := this.commandsMinCursor()
    if len(this.commandsSources) == 1 {
        this.cursorX = minCX
        return
    }

    // Log.Info("dd ", this.cursorY, this.commandsShowBegin, len(this.commandsSources))

    maxCY := this.commandsMaxCursorY()
    if this.cursorY == maxCY {
        if this.commandsShowBegin > 0 {
            this.commandsShowBegin--
        } else {
            this.cursorY--
        }
    }
    this.cursorX = minCX
    return
}
func (this *Terminal) commandsSourceCurrentLine() string {
    return this.commandsSources[this.commandsSourceCurrentLinePosition()]
}

func (this *Terminal) commandsSourceCurrentLinePosition() int {
    return this.cursorY + this.commandsShowBegin
}

func (this *Terminal) commandsChangeMode(mode Mode) {
    this.commandsMode = mode
    if mode == ModeInsert {
        this.commandsBottomContent = "-- INSERT --"
    }

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

func (this *Terminal) framesChangeByBackspace() {
    preWord := this.commandsPreWord()
    cx, _ := this.commandsCursor()
    isHideTablesFrames := isHideTablesFrames(
        this.commandsSourceCurrentLine(), cx,
    )
    Log.Info("isShowFrames ", this.isShowFrames)
    if this.isShowFrames {
        if isHideTablesFrames {
            this.isShowFrames = false
            return
        }

        filter := preWord
        if this.commandsPreRune() == ' ' {
            filter = ""
        }
        this.framesInitForTables(filter)
    }

}
func (this *Terminal) framesChangeByInsert() {
    preWord := this.commandsPreWord()
    cx, _ := this.commandsCursor()
    isShowTablesFrames := isShowTablesFrames(
        this.commandsSourceCurrentLine(), cx,
    )

    if isShowTablesFrames && !this.isShowFrames {
        this.framesInitForTables("")
        this.framesPositionX = this.cursorX - 1
        return
    }

    if this.e.ch == ' ' && this.isShowFrames {
        this.isShowFrames = false
        return
    }

    if this.isShowFrames {
        Log.Infof("preword %s", preWord)
        this.framesInitForTables(preWord)
        return
    }

    if this.e.ch == ';' && this.isShowFrames {
        this.isShowFrames = false
        return
    }

}

func (this *Terminal) framesInitForTables(filter string) {
    this.isShowFrames = true
    this.framesMode = FramesModeTables
    this.frames = arrayFilterLikeString(this.tables[1:], filter)
    this.framesHighlightLinePosition = -1
    this.framesShowBegin = 0

    _, maxLength := arrayMaxLength(this.frames)
    this.framesWidth = maxLength + 3
}

func (this *Terminal) framesInitForResultsDetail() {
    fields := this.resultsCurrentLine()
    columns := this.resultsColumns

    this.isShowFrames = true
    this.framesMode = FramesModeResultsDetail
    frames := make([]string, 0)
    for i, d := range columns {
        field := fields[i]
        if strings.Contains(field, "{") {
            var jsonData map[string]interface{}
            err := json.Unmarshal([]byte(field), &jsonData)
            if err == nil {
                bytes, err := json.MarshalIndent(jsonData, "", "  ")
                jsonStr := string(bytes)
                if err == nil {
                    jsonArr := strings.Split(jsonStr, "\n")
                    for j, ja := range jsonArr {
                        fmt := strings.Repeat(" ", len(d)) + "  "
                        if j == 0 {
                            fmt = d + ": "
                        }
                        frames = append(frames, fmt + ja)
                    }

                }
            }
        } else {

            frames = append(frames, fmt.Sprintf("%s: %s", d, field))
        }
    }
    this.frames = frames
    this.framesPositionX = this.cursorX
    this.framesHighlightLinePosition = -1
    this.framesShowBegin = 0
    _, maxLength := arrayMaxLength(this.frames)
    maxLength = maxLength + 2
    resutsWidth, _ := this.resultsSize()
    this.framesWidth = min(maxLength, resutsWidth)
}

func (this *Terminal) framesMoveUp() {
    if this.framesMode == FramesModeResultsDetail {
        this.framesShowBegin -= this.framesHeight / 2
        if this.framesShowBegin < 0 {
            this.framesShowBegin = 0
        }
        return
    }
    if len(this.frames) == this.framesHeight {
        if this.framesHighlightLinePosition <= 0 {
            this.framesHighlightLinePosition = this.framesHeight - 1
        } else {
            this.framesHighlightLinePosition--
        }
        return
    }

    if this.framesHighlightLinePosition <= 0 {
        if this.framesShowBegin > 0 {
            this.framesShowBegin--
        } else {
            this.framesHighlightLinePosition = this.framesHeight - 1
            this.framesShowBegin = len(this.frames) - this.framesHeight
        }
    } else {
        this.framesHighlightLinePosition--
    }
}

func (this *Terminal) framesMoveDown() {
    if this.framesMode == FramesModeResultsDetail {
        wantShowBegin := this.framesShowBegin + this.framesHeight / 2
        if wantShowBegin < len(this.frames) {
            this.framesShowBegin = wantShowBegin
        }
        return
    }
    offset := 1
    if len(this.frames) == this.framesHeight {
        if this.framesHighlightLinePosition < this.framesHeight - 1{
            this.framesHighlightLinePosition += offset
        } else {
            this.framesHighlightLinePosition=0
        }
        return
    }

    if this.framesShowBegin == len(this.frames) - this.framesHeight {
        this.framesShowBegin = 0
        this.framesHighlightLinePosition=0
        return
    }

    if this.framesHighlightLinePosition < this.framesHeight - 1 {
        this.framesHighlightLinePosition += offset
    } else {
        Log.Infof(
            "frames %d framheight %d showb %d",
            this.frames, this.framesHeight, this.framesShowBegin,
        )
        if len(this.frames) - this.framesHeight  > this.framesShowBegin {
            this.framesShowBegin += offset
        }
    }

}

func (this *Terminal) resultsMoveRight()  {
    width := this.resultsFormatWidth()
    if width == 0 {
        return
    }

    resultWidth, _ := this.resultsSize()

    if this.resultsLeftShowBegin >= width - resultWidth {
        return
    }

    this.resultsLeftShowBegin += resultWidth / 2
}
func (this *Terminal) resultsMoveLeft()  {

    resultWidth, _ := this.resultsSize()

    if this.resultsLeftShowBegin > 0 {
        offset := min(resultWidth / 2, this.resultsLeftShowBegin)
        this.resultsLeftShowBegin -= offset
    }
}

func (this *Terminal) resultsFormatWidth() (width int) {
    width = 0
    if len(this.results) <= 1 {
        return
    }

    width = len(this.resultsFormat[0])
    return
}
func (this *Terminal) resultsSize() (int, int) {
    x := this.width - this.tableSplitSymbolPosition - 2
    y := this.height - this.resultsSplitSymbolPosition - 2
    return x, y
}
func (this *Terminal) resultsPosition() (x, y int) {
    return this.tableSplitSymbolPosition + 1, this.resultsSplitSymbolPosition + 1
}

func (this *Terminal) resultsCurrentLine() (res []string) {
    res = make([]string, 0)

    _, cy := this.resultsCursor()

    index := cy / 2 + this.resultsShowBegin / 2 + 1
    res = this.results[index]

    return
}
func (this *Terminal) resultsCursor() (x int, y int) {
    x = this.cursorX - this.tableSplitSymbolPosition - 1
    y = this.cursorY - this.resultsSplitSymbolPosition - 3
    return
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
    this.isShowFrames = false
}

func (this *Terminal) moveCursorToTables() {
    if this.position == PositionCommands {
        this.commandsLastCursorX = this.cursorX
        this.commandsLastCursorY = this.cursorY
    }
    this.cursorX = 0
    this.cursorY = this.tablesLastCursorY
    this.position = PositionTables
    this.isShowFrames = false
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
    this.isShowFrames = false
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
            if len(this.results) == 1 {
                return
            }

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



