package terminal

import (
	"fmt"
	"io"
)

type Terminal struct {
	stdin            io.WriteCloser
	title            string
	staticTitle      string
	TitleUpdate      bool
	rows             int
	columns          int
	eState           EscapeState
	style            Style
	cursor           Cursor
	cursorMemory     Cursor
	cursorHidden     bool
	altScreenEnabled bool
	screen           *Screen
	altScreen        *Screen
}

func NewTerminal(stdin io.WriteCloser, title string) *Terminal {
	term := &Terminal{
		stdin:            stdin,
		title:            title,
		staticTitle:      "",
		TitleUpdate:      false,
		rows:             40,
		columns:          80,
		eState:           NewEscapeState(),
		style:            NewStyle(),
		cursor:           Cursor{x: 1, y: 1},
		cursorMemory:     Cursor{x: 1, y: 1},
		cursorHidden:     false,
		altScreenEnabled: false,
	}
	term.screen = &Screen{term: term}
	term.altScreen = &Screen{term: term}
	return term
}

func (t *Terminal) ProcessCharacter(r rune) {
	screen := t.GetScreen()
	if t.ProcessEscape(r) {
		return
	}
	switch r {
	case '\r':
		t.cursor.x = 1
	case '\b':
		if t.cursor.x > 1 {
			t.cursor.x--
		}
	case '\a':
		fmt.Println("BELL")
	case '\n':
		screen.MoveToNextLine()
	default:
		active_row := screen.GetCurrentRow()
		if t.cursor.x > t.columns { // full row
			active_row = screen.MoveToNextLine()
		}
		active_row.AddText(r, t.cursor.x, &t.style)
		t.cursor.x++
	}
	screen.Truncate(500)
}

func (t *Terminal) GetScreen() *Screen {
	if t.altScreenEnabled {
		return t.altScreen
	}
	return t.screen
}

func (t *Terminal) SetSize(rows, cols int) bool {
	if rows != t.rows || cols != t.columns {
		if t.cursor.y > rows {
			t.cursor.y = rows
		}
		if len(t.GetScreen().buffor) >= rows && rows > t.rows {
			t.cursor.y += rows - t.rows
		}
		t.rows = rows
		t.columns = cols
		return true
	}
	return false
}

func (t *Terminal) String() string {
	screen := t.GetScreen()
	return screen.String()
}

func (t *Terminal) Bytes() []byte {
	return []byte(t.String())
}

func (t *Terminal) Title() string {
	if len(t.staticTitle) == 0 {
		return t.title
	}
	return t.staticTitle
}

func (t *Terminal) SetTitle(title string) {
	if t.title != title {
		t.title = title
		if len(t.staticTitle) == 0 {
			t.TitleUpdate = true
		}
	}
}

func (t *Terminal) SetStaticTitle(title string) {
	t.staticTitle = title
}

func (t *Terminal) ClearScreen(mode int) {
	screen := t.GetScreen()
	screen.Clear(mode)
}

func (t *Terminal) ClearLine(mode int) {
	fmt.Println("ClearLine", mode)
	currentRow := t.GetScreen().GetCurrentRow()
	switch mode {
	case 0: // from cursor to the end
		currentRow.ClearToEnd(t.cursor.x)
	case 1: // from cursor to the beginning

	case 2: // entire line
		currentRow.Clear()
	}
}

func (t *Terminal) AddStyle(style Style) {
	if t.style != style {
		// screen := t.GetScreen()
		// row := screen.GetCurrentRow()
		if !t.style.IsEmpty() {
			// row.AddStyleEnd()
		}
		if !style.IsEmpty() {
			// row.AddStyle(style)
		}
		t.style = style
	}
}

func (t *Terminal) SaveCursor() {
	t.cursorMemory = t.cursor
}

func (t *Terminal) RestoreCursor() {
	t.cursor = t.cursorMemory
}
