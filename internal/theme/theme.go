package theme

import (
	"fmt"
	"strconv"
)

type Color struct {
	R uint8
	G uint8
	B uint8
}

func (c *Color) String() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

type Theme struct {
	Name         string `json:"name"`
	Foreground   Color  `json:"foreground"`
	Background   Color  `json:"background"`
	Black        Color  `json:"black"`
	Red          Color  `json:"red"`
	Green        Color  `json:"green"`
	Yellow       Color  `json:"yellow"`
	Blue         Color  `json:"blue"`
	Purple       Color  `json:"purple"`
	Cyan         Color  `json:"cyan"`
	White        Color  `json:"white"`
	BrightBlack  Color  `json:"brightBlack"`
	BrightRed    Color  `json:"brightRed"`
	BrightGreen  Color  `json:"brightGreen"`
	BrightYellow Color  `json:"brightYellow"`
	BrightBlue   Color  `json:"brightBlue"`
	BrightPurple Color  `json:"brightPurple"`
	BrightCyan   Color  `json:"brightCyan"`
	BrightWhite  Color  `json:"brightWhite"`
}

func (f *Color) UnmarshalJSON(b []byte) error {
	if len(b) != 9 {
		// If the input data is empty, let's do nothing and throw an error
		return fmt.Errorf("wrong color size")
	} else if b[0] == '"' && b[1] == '#' {
		val, err := strconv.ParseUint(string(b[2:4]), 16, 8)
		if err != nil {
			return err
		}
		f.R = uint8(val)
		val, err = strconv.ParseUint(string(b[4:6]), 16, 8)
		if err != nil {
			return err
		}
		f.G = uint8(val)
		val, err = strconv.ParseUint(string(b[6:8]), 16, 8)
		if err != nil {
			return err
		}
		f.B = uint8(val)
		return nil
	} else {
		return fmt.Errorf("it is not a color")
	}
}
