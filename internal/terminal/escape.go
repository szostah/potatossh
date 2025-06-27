package terminal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	esc      = '\u001b'
	csi      = '[' // Control Sequence Introducer
	osc      = ']' // Operating System Command (ESC]0;this is the window title BEL)
	charset  = '('
	decsc    = '7'
	decrc    = '8'
	bel      = '\a'
	lf       = '\n'
	cr       = '\r'
	bs       = '\b'
	csiFinal = "@[\\]^_`{|}~"
)

type EscapeCode interface {
	IsFinished(r rune) bool
	Parse(s string)
	Execute(term *Terminal)
}

type CSI struct {
	private bool
	command byte
	args    []int
}
type OSC struct {
	command int
	content string
}
type G0CharSet struct {
}

type csiAction func(term *Terminal, args []int)
type csiPrivateAction func(term *Terminal, enable bool)

var csiActions = map[byte]csiAction{
	'm': func(term *Terminal, args []int) {
		style := term.style
		if len(args) == 0 {
			style = style.Add(0)
		}
		style = style.AddStyles(args)
		term.AddStyle(style)
	},

	'J': func(term *Terminal, args []int) {
		if len(args) == 0 {
			term.ClearScreen(0)
		} else {
			term.ClearScreen(args[0])
		}
	},
	'G': func(term *Terminal, args []int) { //CHA
		if len(args) == 0 {
			term.cursor.x = 1
		} else {
			term.cursor.x = args[0]
			if term.cursor.x > term.columns {
				term.cursor.x = term.columns
			}
		}
	},
	'd': func(term *Terminal, args []int) { // VPA
		if len(args) == 0 {
			term.cursor.y = 1
		} else {
			term.cursor.y = args[0]
			if term.cursor.y > term.rows {
				term.cursor.y = term.rows
			}
		}
	},
	'H': func(term *Terminal, args []int) {
		if len(args) == 0 {
			term.cursor.y = 1
			term.cursor.x = 1
		} else if len(args) == 1 {
			term.cursor.y = args[0]
			if term.cursor.y > term.rows {
				term.cursor.y = term.rows
			}
			term.cursor.x = 1
		} else {
			term.cursor.y = args[0]
			if term.cursor.y > term.rows {
				term.cursor.y = term.rows
			}
			term.cursor.x = args[1]
			if term.cursor.x > term.columns {
				term.cursor.x = term.columns
			}
		}
	},

	'K': func(term *Terminal, args []int) {
		if len(args) == 0 {
			term.ClearLine(0)
		} else {
			term.ClearLine(args[0])
		}
	},

	'P': func(term *Terminal, args []int) { // 	DCH
		if len(args) == 0 {
			term.GetScreen().GetCurrentRow().RemoveN(term.cursor.x, 1)
		} else {
			term.GetScreen().GetCurrentRow().RemoveN(term.cursor.x, args[0])
		}
		// term.ClearLine(0) // TODO Delete Character
	},

	'X': func(term *Terminal, args []int) { // ECH TODO
		if len(args) == 0 {
			term.GetScreen().GetCurrentRow().EraseToN(term.cursor.x, 1)
		} else {
			term.GetScreen().GetCurrentRow().EraseToN(term.cursor.x, args[0])
		}
	},

	'A': func(term *Terminal, args []int) {
		if len(args) == 0 {
			if term.cursor.y > 1 {
				term.cursor.y--
			}
		} else {
			term.cursor.y -= args[0]
			if term.cursor.y < 1 {
				term.cursor.y = 1
			}
		}
	},

	'B': func(term *Terminal, args []int) {
		if len(args) == 0 {
			if term.cursor.y < term.rows {
				term.cursor.y++
			}
		} else {
			term.cursor.y += args[0]
			if term.cursor.y > term.rows {
				term.cursor.y = term.rows
			}
		}
	},

	'C': func(term *Terminal, args []int) {
		if len(args) == 0 {
			if term.cursor.x < term.columns {
				term.cursor.x++
			}
		} else {
			term.cursor.x += args[0]
			if term.cursor.x > term.columns {
				term.cursor.x = term.columns
			}
		}
	},

	'D': func(term *Terminal, args []int) {
		if len(args) == 0 {
			if term.cursor.x > 1 {
				term.cursor.x--
			}
		} else {
			term.cursor.x -= args[0]
			if term.cursor.x < 1 {
				term.cursor.x = 1
			}
		}
	},

	'n': func(term *Terminal, args []int) { // request cursor position
		if len(args) == 1 && args[0] == 6 {
			// CSI r ; c R
			fmt.Println("Sending cursor position back!")
			term.stdin.Write([]byte(fmt.Sprintf("%c[%d;%dR", esc, term.cursor.y, term.cursor.x)))
		}
	},
	'@': func(term *Terminal, args []int) {
		n := 1
		if len(args) == 1 {
			n = args[0]
		}
		style := NewStyle()
		for i := range n {
			term.GetScreen().GetCurrentRow().InsertText(' ', term.cursor.x+i, &style)
		}
	},
}

var csiPrivateActions = map[int]csiPrivateAction{
	1: func(term *Terminal, enable bool) {
		fmt.Println("DECCKM:", enable)
	},
	4: func(term *Terminal, enable bool) {
		fmt.Println("Smooth scroll:", enable)
	},
	12: func(term *Terminal, enable bool) {
		fmt.Println("Blinking cursor:", enable)
	},
	25: func(term *Terminal, enable bool) {
		term.cursorHidden = !enable
	},
	1004: func(term *Terminal, enable bool) {
		fmt.Println("Reporting focus:", enable)
	},
	1049: func(term *Terminal, enable bool) {
		fmt.Println("Alt screen:", enable)
		if enable {
			term.SaveCursor()
		} else {
			term.RestoreCursor()
		}
		term.altScreenEnabled = enable
	},
	2004: func(term *Terminal, enable bool) {
		fmt.Println("Bracketed paste mode:", enable)
	},
}

func (_ CSI) IsFinished(r rune) bool {
	return unicode.IsLetter(r) || strings.ContainsRune(csiFinal, r)
}

func (ec *G0CharSet) Parse(s string) {} // Ignore for now

func (ec *G0CharSet) Execute(term *Terminal) {} // Ignore for now

func (_ G0CharSet) IsFinished(r rune) bool {
	/*
		ESC (     Start sequence defining G0 character set
		ESC ( B   Select default (ISO 8859-1 mapping)
		ESC ( 0   Select VT100 graphics mapping
		ESC ( U   Select null mapping - straight to character ROM
		ESC ( K   Select user mapping - the map that is loaded by the utility mapscrn(8).
	*/
	return r == 'B' || r == '0' || r == 'U' || r == 'K'
}

func (ec *CSI) Parse(s string) {
	ec.command = s[len(s)-1]
	if len(s) > 1 {
		s = s[:len(s)-1]

		if s[0] == '?' {
			ec.private = true
			s = s[1:]
		}

		for _, n := range strings.Split(s, ";") {
			p, err := strconv.Atoi(n)
			if err != nil {

			}
			ec.args = append(ec.args, p)
		}
	}
}

func (ec *CSI) Execute(term *Terminal) {
	if ec.private {
		if len(ec.args) == 1 {
			f, ok := csiPrivateActions[ec.args[0]]
			if ok {
				f(term, ec.command == 'h')
			} else {
				fmt.Println("CSI private", ec.args[0], "not implemented", ec.command == 'h')
			}
		}
	} else {
		f, ok := csiActions[ec.command]
		if ok {
			f(term, ec.args)
		} else {
			fmt.Println("CSI", string(ec.command), ec.args, "not implemented")
		}
	}
}

func (OSC) IsFinished(r rune) bool {
	return r == bel
}

func (ec *OSC) Parse(s string) {
	elems := strings.Split(s[:len(s)-1], ";")
	var err error
	ec.command, err = strconv.Atoi(elems[0])
	if err != nil {

	}
	ec.content = elems[1]
}

func (ec *OSC) Execute(term *Terminal) {
	if ec.command == 0 {
		term.SetTitle(ec.content)
	}
}

type EscapeState struct {
	enabled    bool
	escapeCode EscapeCode
	buffor     string
}

func NewEscapeState() EscapeState {
	return EscapeState{enabled: false, escapeCode: nil, buffor: ""}
}

func (t *Terminal) ProcessEscape(r rune) bool {
	if !t.eState.enabled {
		if r == esc {
			t.eState.enabled = true
			return true
		}
	} else if t.eState.escapeCode == nil {
		switch r {
		case csi:
			t.eState.escapeCode = &CSI{}
		case osc:
			t.eState.escapeCode = &OSC{}
		case charset:
			t.eState.escapeCode = &G0CharSet{}
		case decsc:
			t.SaveCursor()
			t.eState.enabled = false
		case decrc:
			t.RestoreCursor()
			t.eState.enabled = false
		default:
			fmt.Println("Unknown escape mode:", string(r), int(r))
			t.eState.enabled = false
		}
		return true
	} else {
		t.eState.buffor += string(r)
		if t.eState.escapeCode.IsFinished(r) {
			t.eState.escapeCode.Parse(t.eState.buffor)
			t.eState.escapeCode.Execute(t)
			t.eState.buffor = ""
			t.eState.enabled = false
			t.eState.escapeCode = nil
		}
		return true
	}

	return false
}
