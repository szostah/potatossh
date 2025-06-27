package terminal

import (
	"fmt"
	"strings"
)

/* SGR ESC[ num m

| color    | fg | bg  |
+----------+----+-----+
| black    | 30 | 40  |
| red      | 31 | 41  |
| green    | 32 | 42  |
| yellow   | 33 | 43  |
| blue     | 34 | 44  |
| magenta  | 35 | 45  |
| cyan     | 36 | 46  |
| white    | 37 | 47  |
| gray     | 90 | 100 |
| bred     | 91 | 101 |
| bgreen   | 92 | 102 |
| byellow  | 93 | 103 |
| bblue    | 94 | 104 |
| bmagenta | 95 | 105 |
| bcyan    | 96 | 106 |
| bwhite   | 97 | 107 |

*/

var colors = [8]string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
}

var values216 = [6]int{
	0, 51, 102, 153, 204, 255,
}

type RgbColor struct {
	r, g, b int
}

type Style struct {
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	blink         bool
	invert        bool
	strike        bool
	brightFgColor bool
	brightBgColor bool
	fgColor       int
	bgColor       int
	rgbFgColor    *RgbColor
	rgbBgColor    *RgbColor
}

func NewStyle() Style {
	return Style{fgColor: -1, bgColor: -1, rgbFgColor: nil, rgbBgColor: nil}
}

func (s *Style) IsEmpty() bool {
	return s.bold == false &&
		s.dim == false &&
		s.italic == false &&
		s.underline == false &&
		s.blink == false &&
		s.invert == false &&
		s.strike == false &&
		s.brightFgColor == false &&
		s.brightBgColor == false &&
		s.fgColor == -1 &&
		s.bgColor == -1 &&
		s.rgbFgColor == nil &&
		s.rgbBgColor == nil
}

func (s Style) Add(sgr int) Style {
	switch {
	case sgr == 0:
		return NewStyle()
	case sgr == 1:
		s.bold = true
	case sgr == 2:
		s.dim = true
	case sgr == 3:
		s.italic = true
	case sgr == 4:
		s.underline = true
	case sgr == 5 || sgr == 6:
		s.blink = true
	case sgr == 7:
		s.invert = true
	case sgr == 9:
		s.strike = true
	case sgr == 22:
		s.bold = false
		s.dim = false
	case sgr == 23:
		s.italic = false
	case sgr == 24:
		s.underline = false
	case sgr == 25:
		s.blink = false
	case sgr == 27:
		s.invert = false
	case sgr >= 30 && sgr <= 37:
		s.rgbFgColor = nil
		s.brightFgColor = false
		s.fgColor = sgr - 30
	case sgr == 39:
		s.rgbFgColor = nil
		s.brightFgColor = false
		s.fgColor = -1
	case sgr >= 40 && sgr <= 47:
		s.rgbBgColor = nil
		s.brightFgColor = false
		s.bgColor = sgr - 40
	case sgr == 49:
		s.rgbBgColor = nil
		s.brightBgColor = false
		s.bgColor = -1
	case sgr >= 90 && sgr <= 97:
		s.rgbFgColor = nil
		s.brightFgColor = true
		s.fgColor = sgr - 90
	case sgr >= 100 && sgr <= 107:
		s.rgbBgColor = nil
		s.brightBgColor = true
		s.bgColor = sgr - 100
	}
	return s
}

func (s Style) Add8BitColor(n int, bg bool) Style {

	if n < 0 {
		return s
	}

	if bg {
		switch {
		case n < 8:
			s.bgColor = n
		case n < 16:
			s.bgColor = n - 8
			s.brightBgColor = true
		case n < 232:
			color_index := n - 16
			red_index := color_index / 36
			green_index := (color_index % 36) / 6
			blue_index := color_index % 6
			s.rgbBgColor = &RgbColor{r: values216[red_index], g: values216[green_index], b: values216[blue_index]}
		case n < 256:
			value := (n-232)*10 + 8
			s.rgbBgColor = &RgbColor{r: value, g: value, b: value}
		}
	} else {
		switch {
		case n < 8:
			s.fgColor = n
		case n < 16:
			s.fgColor = n - 8
			s.brightFgColor = true
		case n < 232:
			color_index := n - 16
			red_index := color_index / 36
			green_index := (color_index % 36) / 6
			blue_index := color_index % 6
			s.rgbFgColor = &RgbColor{r: values216[red_index], g: values216[green_index], b: values216[blue_index]}
		case n < 256:
			value := (n-232)*10 + 8
			s.rgbFgColor = &RgbColor{r: value, g: value, b: value}
		}
	}

	return s
}

const (
	NO_COLOR = iota
	FG_CUBE_COLOR
	BG_CUBE_COLOR
	FG_TRUE_COLOR
	BG_TRUE_COLOR
)

func (s Style) AddStyles(sgrs []int) Style { // 8 bit
	color := NO_COLOR
	rgb := RgbColor{}
	colorCnt := 0
	for _, sgr := range sgrs {
		if color == NO_COLOR {
			if sgr == 38 {
				color = FG_CUBE_COLOR
			} else if sgr == 48 {
				color = BG_CUBE_COLOR
			} else {
				s = s.Add(sgr)
			}
		} else { // extended color (8 or 24 bit)
			switch colorCnt {
			case 0:
				if sgr == 2 {
					if color == FG_CUBE_COLOR {
						color = FG_TRUE_COLOR
					} else {
						color = BG_TRUE_COLOR
					}
				}
			case 1:
				if color == FG_CUBE_COLOR || color == BG_CUBE_COLOR {
					s = s.Add8BitColor(sgr, color == BG_CUBE_COLOR)
					color = NO_COLOR
					colorCnt = 0
					continue
				} else {
					rgb.r = sgr
				}
			case 2:
				rgb.g = sgr
			case 3:
				rgb.b = sgr
				if color == FG_TRUE_COLOR {
					s.rgbFgColor = &rgb
				} else {
					s.rgbBgColor = &rgb
				}
				color = NO_COLOR
				colorCnt = 0
				continue
			}
			colorCnt++
		}

	}
	return s
}

func (s *Style) Attributes() string {
	html := ""
	classes := []string{}
	if s.bold {
		classes = append(classes, "bold")
	}
	if s.dim {
		classes = append(classes, "dim")
	}
	if s.italic {
		classes = append(classes, "italic")
	}
	if s.underline {
		classes = append(classes, "underline")
	}
	if s.blink {
		classes = append(classes, "blink")
	}
	if s.invert {
		classes = append(classes, "invert")
	}
	if s.strike {
		classes = append(classes, "strike")
	}
	if s.brightFgColor {
		classes = append(classes, "fg_bright")
	}
	if s.fgColor != -1 {
		classes = append(classes, "fg_"+colors[s.fgColor])
	}
	if s.brightBgColor {
		classes = append(classes, "bg_bright")
	}
	if s.bgColor != -1 {
		classes = append(classes, "bg_"+colors[s.bgColor])
	}
	if len(classes) > 0 {
		html = "class=\"" + strings.Join(classes, " ") + "\""
	}

	if s.rgbFgColor != nil || s.rgbBgColor != nil {
		if len(html) > 0 {
			html += " "
		}
		html += "style=\""
		if s.rgbFgColor != nil {
			html += fmt.Sprintf("color: rgb(%d,%d,%d);", s.rgbFgColor.r, s.rgbFgColor.g, s.rgbFgColor.b)
		}
		if s.rgbBgColor != nil {
			html += fmt.Sprintf("background-color: rgb(%d,%d,%d);", s.rgbBgColor.r, s.rgbBgColor.g, s.rgbBgColor.b)
		}
		html += "\""
	}

	return html
}

func (s *Style) Classes() string {
	out := []string{}
	if s.bold {
		out = append(out, "bold")
	}
	if s.dim {
		out = append(out, "dim")
	}
	if s.italic {
		out = append(out, "italic")
	}
	if s.underline {
		out = append(out, "underline")
	}
	if s.blink {
		out = append(out, "blink")
	}
	if s.invert {
		out = append(out, "invert")
	}
	if s.strike {
		out = append(out, "strike")
	}
	if s.brightFgColor {
		out = append(out, "fg_bright")
	}
	if s.brightBgColor {
		out = append(out, "bg_bright")
	}
	if s.fgColor != -1 {
		out = append(out, "fg_"+colors[s.fgColor])
	}
	if s.bgColor != -1 {
		out = append(out, "bg_"+colors[s.bgColor])
	}

	return strings.Join(out, " ")
}
