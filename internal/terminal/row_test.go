package terminal

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAddTextAtEnd(t *testing.T) {
	row := Row{}
	style := NewStyle()
	row.AddText('l', 1, &style)
	row.AddText('o', 2, &style)
	row.AddText('r', 3, &style)
	row.AddText('e', 4, &style)
	row.AddText('m', 5, &style)

	result := row.Html()
	want := "lorem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextAtStart(t *testing.T) {
	row := Row{}
	style := NewStyle()
	row.AddText('l', 1, &style)
	row.AddText('o', 2, &style)
	row.AddText('r', 3, &style)
	row.AddText('e', 4, &style)
	row.AddText('m', 5, &style)

	row.AddText('L', 1, &style)

	result := row.Html()
	want := "Lorem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextInMiddle(t *testing.T) {
	row := Row{}
	style := NewStyle()
	row.AddText('l', 1, &style)
	row.AddText('o', 2, &style)
	row.AddText('r', 3, &style)
	row.AddText('e', 4, &style)
	row.AddText('m', 5, &style)

	row.AddText('R', 3, &style)

	result := row.Html()
	want := "loRem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextTwoStyles(t *testing.T) {
	row := Row{}
	style1 := NewStyle()
	style2 := NewStyle()
	style2.bold = true
	row.AddText('l', 1, &style1)
	row.AddText('o', 2, &style1)
	row.AddText('r', 3, &style1)
	row.AddText('e', 4, &style2)
	row.AddText('m', 5, &style2)

	result := row.Html()
	want := "lor<span class=\"bold\">em</span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextTwoStylesMixed(t *testing.T) {
	row := Row{}
	style1 := NewStyle()
	style2 := NewStyle()
	style2.bold = true
	row.AddText('l', 1, &style1)
	row.AddText('o', 2, &style1)
	row.AddText('r', 3, &style2)
	row.AddText('e', 4, &style1)
	row.AddText('m', 5, &style1)

	result := row.Html()
	want := "lo<span class=\"bold\">r</span>em"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextManyStyles(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	red := NewStyle().Add(31)
	green := NewStyle().Add(32)
	cyan := NewStyle().Add(36)

	row.AddText('l', 1, &red)
	row.AddText('o', 2, &bold)
	row.AddText('r', 3, &green)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &cyan)

	result := row.Html()
	want := "<span class=\"fg_red\">l</span><span class=\"bold\">o</span><span class=\"fg_green\">r</span>e<span class=\"fg_cyan\">m</span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleAtStart(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('L', 1, &bold)

	result := row.Html()
	want := "<span class=\"bold\">L</span>orem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleAtEnd(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('M', 5, &bold)

	result := row.Html()
	want := "lore<span class=\"bold\">M</span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleInMiddle(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('R', 3, &bold)

	result := row.Html()
	want := "lo<span class=\"bold\">R</span>em"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleFirstTwo(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('L', 1, &bold)
	row.AddText('O', 2, &bold)

	result := row.Html()
	want := "<span class=\"bold\">LO</span>rem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleFirstTwoReverset(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('O', 2, &bold)
	row.AddText('L', 1, &bold)

	result := row.Html()
	want := "<span class=\"bold\">LO</span>rem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestAddTextUpdateStyleFirstTwoMixed(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)
	red := NewStyle().Add(31)

	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.AddText('O', 2, &bold)
	row.AddText('L', 1, &red)

	result := row.Html()
	want := "<span class=\"fg_red\">L</span><span class=\"bold\">O</span>rem"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
	fmt.Println("RESULT", result)
}

func TestAddTextUpdateStyleTheLast(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)

	row.AddText('l', 1, &bold)
	row.AddText('o', 2, &bold)
	row.AddText('r', 3, &bold)
	row.AddText('e', 4, &bold)
	row.AddText('m', 5, &reset)

	row.AddText('M', 5, &bold)

	result := row.Html()
	want := "<span class=\"bold\">loreM</span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
	fmt.Println("RESULT", result)
}

func TestAddTextUpdateStyleInMiddle2(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)

	row.AddText('l', 1, &bold)
	row.AddText('o', 2, &bold)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &bold)
	row.AddText('m', 5, &bold)

	row.AddText('R', 3, &bold)

	result := row.Html()
	want := "<span class=\"bold\">loRem</span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestCursorWithSpaces(t *testing.T) {
	row := Row{}
	reset := NewStyle()

	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	result := row.HtmlWithCursor(10)
	want := "lorem    <span class=\"cursor\"> </span>"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestUtf8WithStyle(t *testing.T) {
	row := Row{}
	reset := NewStyle()

	row.AddText('ż', 1, &reset)
	row.AddText('ó', 2, &reset)
	row.AddText('ł', 3, &reset)
	row.AddText('ć', 4, &reset)

	result := row.Html()
	want := "żółć"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
	fmt.Println("RESULT", result)
}

func TestClearToEnd(t *testing.T) {
	row := Row{}
	reset := NewStyle()
	bold := NewStyle().Add(1)

	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &bold)

	row.ClearToEnd(3)

	result := row.Html()
	want := "lo"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
}

func TestRemoveN(t *testing.T) {
	style1 := NewStyle()
	style2 := NewStyle().Add(1)
	style3 := NewStyle().Add(2)

	t.Run("Empty row", func(t *testing.T) {
		row := Row{}
		row.RemoveN(1, 1)
		expectedRow := Row{}
		if !reflect.DeepEqual(row, expectedRow) {
			t.Errorf("Oczekiwano: %v, otrzymano: %v", expectedRow, row)
		}
	})
	t.Run("Out of range", func(t *testing.T) {
		row := Row{text: []rune("ABC")}
		row.RemoveN(4, 1) // Pozycja poza zakresem
		expectedRow := Row{text: []rune("ABC")}
		if !reflect.DeepEqual(row, Row{text: []rune("ABC")}) {
			t.Errorf("Oczekiwano: %v, otrzymano: %v", expectedRow, row)
		}
	})
	t.Run("Negative N", func(t *testing.T) {
		row := Row{text: []rune("ABC")}
		row.RemoveN(1, -1) // Ujemna liczba znaków
		expectedRow := Row{text: []rune("ABC")}
		if !reflect.DeepEqual(row, Row{text: []rune("ABC")}) {
			t.Errorf("Oczekiwano: %v, otrzymano: %v", expectedRow, row)
		}
	})
	t.Run("Middle of one argument", func(t *testing.T) {
		row := Row{
			text:  []rune("ABCDEF"),
			attrs: []Attr{{start: 1, end: 6, style: style1}},
		}
		row.RemoveN(2, 2)
		expectedText := []rune("ADEF")
		expectedAttrs := []Attr{{start: 1, end: 4, style: style1}}
		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany")
		}
	})
	t.Run("Whole attribute", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHIJ"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 10, style: style3},
			},
		}
		row.RemoveN(4, 3)
		expectedText := []rune("ABCGHIJ")
		expectedAttrs := []Attr{
			{start: 1, end: 3, style: style1},
			{start: 4, end: 7, style: style3},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})
	t.Run("Whole attribute surrounded by the same", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHIJ"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 10, style: style1},
			},
		}
		row.RemoveN(4, 3)
		expectedText := []rune("ABCGHIJ")
		expectedAttrs := []Attr{
			{start: 1, end: 7, style: style1},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})
	t.Run("Start of one argument", func(t *testing.T) {
		row := Row{
			text:  []rune("ABCDEF"),
			attrs: []Attr{{start: 1, end: 6, style: style1}},
		}
		row.RemoveN(1, 3)
		expectedText := []rune("DEF")
		expectedAttrs := []Attr{{start: 1, end: 3, style: style1}}
		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany")
		}
	})
	t.Run("End of one argument", func(t *testing.T) {
		row := Row{
			text:  []rune("ABCDEF"),
			attrs: []Attr{{start: 1, end: 6, style: style1}},
		}
		row.RemoveN(4, 3)
		expectedText := []rune("ABC")
		expectedAttrs := []Attr{{start: 1, end: 3, style: style1}}
		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany")
		}
	})

	t.Run("Two whole attributes in middle", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGH"),
			attrs: []Attr{
				{start: 1, end: 2, style: style1},
				{start: 3, end: 4, style: style2},
				{start: 5, end: 6, style: style3},
				{start: 7, end: 8, style: style2},
			},
		}
		row.RemoveN(3, 4)
		expectedText := []rune("ABGH")
		expectedAttrs := []Attr{
			{start: 1, end: 2, style: style1},
			{start: 3, end: 4, style: style2},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("Two whole attributes at start", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style3},
			},
		}
		row.RemoveN(1, 6)
		expectedText := []rune("GHI")
		expectedAttrs := []Attr{
			{start: 1, end: 3, style: style3},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("Two whole attributes at end", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style3},
			},
		}
		row.RemoveN(4, 6)
		expectedText := []rune("ABC")
		expectedAttrs := []Attr{
			{start: 1, end: 3, style: style1},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("1.5 attributes at start", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style3},
			},
		}
		row.RemoveN(1, 4)
		expectedText := []rune("EFGHI")
		expectedAttrs := []Attr{
			{start: 1, end: 2, style: style2},
			{start: 3, end: 5, style: style3},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})
	t.Run("1.5 attributes at end", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style3},
			},
		}
		row.RemoveN(3, 4)
		expectedText := []rune("ABGHI")
		expectedAttrs := []Attr{
			{start: 1, end: 2, style: style1},
			{start: 3, end: 5, style: style3},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("1.75 attributes in middle", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style3},
			},
		}
		row.RemoveN(3, 5)
		expectedText := []rune("ABHI")
		expectedAttrs := []Attr{
			{start: 1, end: 2, style: style1},
			{start: 3, end: 4, style: style3},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("Two whole attributes in middle surrounded by the same style", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGH"),
			attrs: []Attr{
				{start: 1, end: 2, style: style1},
				{start: 3, end: 4, style: style2},
				{start: 5, end: 6, style: style3},
				{start: 7, end: 8, style: style1},
			},
		}
		row.RemoveN(3, 4)
		expectedText := []rune("ABGH")
		expectedAttrs := []Attr{
			{start: 1, end: 4, style: style1},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})

	t.Run("1.75 attributes in middle surrounded by the same style", func(t *testing.T) {
		row := Row{
			text: []rune("ABCDEFGHI"),
			attrs: []Attr{
				{start: 1, end: 3, style: style1},
				{start: 4, end: 6, style: style2},
				{start: 7, end: 9, style: style1},
			},
		}
		row.RemoveN(3, 5)
		expectedText := []rune("ABHI")
		expectedAttrs := []Attr{
			{start: 1, end: 4, style: style1},
		}

		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany (wiele atrybutów)")
		}
	})
	t.Run("Remove all row", func(t *testing.T) {
		row := Row{
			text:  []rune("ABCDEF"),
			attrs: []Attr{{start: 1, end: 6, style: style1}},
		}
		row.RemoveN(1, 6)
		expectedText := []rune{}
		expectedAttrs := []Attr{}
		if !reflect.DeepEqual(row.text, expectedText) || !reflect.DeepEqual(row.attrs, expectedAttrs) {
			t.Errorf("Test nieudany")
		}
	})
}

func TestEraseToN(t *testing.T) {
	row := Row{}
	reset := NewStyle()

	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)

	row.EraseToN(2, 2)

	result := row.Html()
	want := "l  em"

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
	fmt.Println("RESULT", result, row.attrs)
}

func TestEraseToNOver(t *testing.T) {
	row := Row{}
	reset := NewStyle()

	row.AddText('l', 1, &reset)
	row.AddText('o', 2, &reset)
	row.AddText('r', 3, &reset)
	row.AddText('e', 4, &reset)
	row.AddText('m', 5, &reset)
	//l
	row.EraseToN(2, 5)

	result := row.Html()
	want := "l     "

	if result != want {
		t.Errorf("Row result: %#q want: %#q", result, want)
	}
	fmt.Println("RESULT", result, row.attrs)
}
