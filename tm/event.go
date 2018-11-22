package tm

import (
    "github.com/nsf/termbox-go"
)

type Event struct {
    preCh rune
    ch rune
    chs []rune
    operatorChs []rune
    operatorIsBegin bool
    preKey termbox.Key
    key termbox.Key
    e termbox.Event
}

func newEvent() (e *Event) {
    e = &Event{
        chs: make([]rune, 0),
        operatorChs: make([]rune, 0),
        operatorIsBegin: false,
    }
    return
}

func (this *Event) clearOperator() {
    this.operatorChs = make([]rune, 0)
    this.operatorIsBegin = false
}

func (this *Event) setCh(r rune) {

    this.preCh = this.ch
    this.ch = r
}

func (this *Event) resetOperator() {

    if inArray(this.ch, []rune("dc")) > -1 {
        this.operatorIsBegin = true
        this.operatorChs = append(this.operatorChs, this.ch)
    } else {
        if this.operatorIsBegin {
            this.operatorChs = append(this.operatorChs, this.ch)
        }
    }
    Log.Infof("operator %s", string(this.operatorChs))
}

