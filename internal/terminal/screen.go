package terminal

type Cursor struct {
	y, x int
}

type Screen struct {
	term   *Terminal
	buffor []Row
}

func (s *Screen) MoveToNextLine() *Row {
	if len(s.buffor) < s.term.rows { // buffor smaller than screen
		if len(s.buffor) == s.term.cursor.y {
			s.buffor = append(s.buffor, Row{})
		}
		s.term.cursor.x = 1
		s.term.cursor.y++
	} else {
		if s.term.rows == s.term.cursor.y {
			s.buffor = append(s.buffor, Row{})
		} else {
			s.term.cursor.y++
		}
		s.term.cursor.x = 1
	}

	return s.GetCurrentRow()
}

func (s *Screen) GetCurrentRow() *Row {
	buffor_size := len(s.buffor)
	if buffor_size == 0 {
		s.buffor = append(s.buffor, Row{})
		return &s.buffor[0]
	}
	if buffor_size > s.term.rows {
		return &s.buffor[(buffor_size-s.term.rows)+s.term.cursor.y-1]
	} else {
		if buffor_size < s.term.cursor.y {
			missingLines := s.term.cursor.y - buffor_size
			for i := 0; i < missingLines; i++ {
				s.buffor = append(s.buffor, Row{})
			}
		}
		return &s.buffor[s.term.cursor.y-1]
	}
}

func (s *Screen) Truncate(rows int) {
	var n int = len(s.buffor) - rows
	if n > 0 {
		s.buffor = s.buffor[n:]
	}
}

func (s *Screen) Clear(mode int) {
	switch mode {
	case 2:
		for i := 0; i < s.term.rows; i++ {
			s.buffor = append(s.buffor, Row{})
		}
	case 3:
		s.buffor = []Row{}
		s.term.cursor.y = 1
		s.term.cursor.x = 1
	}
}

func (s *Screen) String() string {
	out := ""
	active_row := s.GetCurrentRow()
	for i, row := range s.buffor {
		if active_row == &s.buffor[i] {
			if s.term.cursorHidden {
				out += row.Html()
			} else {
				out += row.HtmlWithCursor(s.term.cursor.x)
			}
		} else {
			out += row.Html()
		}
		if i != len(s.buffor)-1 {
			out += "\n"
		}
	}
	return out
}

func (s *Screen) Bytes() []byte {
	return []byte(s.String())
}
