package terminal

import (
	"slices"
	"strings"
)

func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

type Attr struct {
	start int
	end   int
	style Style
}

type Row struct {
	text  []rune
	attrs []Attr
}

func (r *Row) Length() int {
	return len(r.text)
}

func (r *Row) GetAttr(x int) (int, *Attr) {
	if x < 1 || x > r.Length() {
		return -1, nil
	}

	for i, a := range r.attrs {
		if a.start <= x && a.end >= x {
			return i, &r.attrs[i]
		}
	}
	return -1, nil
}

/*
	Case 0: Empty
		No Text == No attributes, add new text + attribute
	Case 1: Cursor ahead text
		Add spaces if needed + text
		If previous attribute doesn't exist or styles differ:
			Add new attribute
		Else:
			Update previous attribute - extend 'end'
	Case 2: Cursor inside text
		Replace text at cursor
		If styles differ:
			2.1 In middle (start<x<end)
				Resize current attribute - shrink 'end' (start->x)
				Add new attribute with new style for cursor (x)
				Copy current attribute (x->end)
			2.2 At beggining (x==start)
				Insert new attribute before previous for cursor
				Resize current attribute - shrink 'start'
			2.3 At end (x==end)
				Resize current attribute - shrink 'end'
				Add new attribute with new style for cursor (x)

*/

// 2
// x
//    y

func (r *Row) AddText(letter rune, x int, s *Style) {
	length := r.Length()
	if x > length {
		_, previousAttr := r.GetAttr(x - 1)
		if x-length-1 > 0 {
			r.text = append(r.text, []rune(strings.Repeat(" ", x-length-1))...)
			r.attrs = append(r.attrs, Attr{start: length + 1, end: x - 1, style: NewStyle()})
		}
		r.text = append(r.text, letter)
		if previousAttr == nil || *s != previousAttr.style {
			r.attrs = append(r.attrs, Attr{start: x, end: x, style: *s})
		} else {
			previousAttr.end = x
		}
	} else {
		i, currentAttr := r.GetAttr(x)
		r.text[x-1] = letter
		if *s != currentAttr.style {
			if currentAttr.start == currentAttr.end { // single attribute
				mergeLeft := i > 0 && r.attrs[i-1].style == *s
				mergeRight := i < len(r.attrs)-1 && r.attrs[i+1].style == *s

				if mergeLeft && mergeRight { // both neighbors the same: left = left+current+right
					r.attrs[i-1].end = r.attrs[i+1].end
					r.attrs = append(r.attrs[:i], r.attrs[i+2:]...)
				} else if mergeLeft { // left neighbor the same: left = left+current
					r.attrs[i-1].end = x
					r.attrs = append(r.attrs[:i], r.attrs[i+1:]...)
				} else if mergeRight { // right neighbor the same: right = current+right
					r.attrs[i+1].start = x
					r.attrs = append(r.attrs[:i], r.attrs[i+1:]...)
				} else { // replace style
					currentAttr.style = *s
				}
			} else if currentAttr.start == x { // left edge
				currentAttr.start = x + 1
				if i > 0 && r.attrs[i-1].style == *s { // merge left
					r.attrs[i-1].end = x
				} else {
					r.attrs = slices.Insert(r.attrs, i, Attr{start: x, end: x, style: *s})
				}
			} else if currentAttr.end == x { // right edge
				currentAttr.end = x - 1
				if i < len(r.attrs)-1 && r.attrs[i+1].style == *s { // merge right
					r.attrs[i+1].start = x
				} else {
					r.attrs = slices.Insert(r.attrs, i+1, Attr{start: x, end: x, style: *s})
				}
			} else { // in the middle
				end := currentAttr.end
				currentAttr.end = x - 1
				r.attrs = slices.Insert(r.attrs, i+1, Attr{start: x, end: x, style: *s}, Attr{start: x + 1, end: end, style: currentAttr.style})
			}
		}
	}
}

func (r *Row) InsertText(letter rune, x int, s *Style) {
	// TODO
	length := r.Length()
	if x > length {
		r.AddText(' ', x, s)
	} else {
		i, currentAttr := r.GetAttr(x)
		r.text = slices.Insert(r.text, x-1, letter)
		currentAttr.end += 1
		for j := range r.attrs[i+1:] {
			r.attrs[i+1+j].start += 1
			r.attrs[i+1+j].end += 1
		}

	}
}

func (r *Row) AddTexts(letters []rune, x int, s *Style) {
	length := r.Length()
	lettersLength := len(letters)

	if x > length {
		if x-length > 0 {
			r.text = append(r.text, []rune(strings.Repeat(" ", x-length))...)
			r.attrs = append(r.attrs, Attr{start: length, end: x - 1, style: NewStyle()})
		}
		r.text = append(r.text, letters...)
		if len(r.attrs) == 0 || *s != r.attrs[len(r.attrs)-1].style {
			r.attrs = append(r.attrs, Attr{start: x, end: x + lettersLength - 1, style: *s})
		} else {
			r.attrs[len(r.attrs)-1].end = x + lettersLength - 1
		}
	} else {
		r.text = append(r.text[:x-1], append(letters, r.text[x-1:]...)...)

		start := x - 1
		end := x + lettersLength - 2

		startIndex, startAttr := r.GetAttr(start)
		endIndex, endAttr := r.GetAttr(end)

		if startAttr != nil && endAttr != nil && *s == startAttr.style && *s == endAttr.style {
			startAttr.end = endAttr.end
			r.attrs = append(r.attrs[:endIndex], r.attrs[endIndex+1:]...)
		} else if startAttr != nil && *s == startAttr.style {
			startAttr.end = end
		} else if endAttr != nil && *s == endAttr.style {
			endAttr.start = start
		} else {
			r.attrs = slices.Insert(r.attrs, startIndex+1, Attr{start: start, end: end, style: *s})
		}

		r.attrs = slices.DeleteFunc(r.attrs, func(a Attr) bool {
			return a.start > end || a.end < start
		})

		r.attrs = slices.CompactFunc(r.attrs, func(a Attr, b Attr) bool {
			return a.style == b.style && a.end+1 == b.start
		})
	}
}

func (r *Row) Clear() {
	r.text = []rune{}
	r.attrs = []Attr{}
}

func (r *Row) ClearToEnd(x int) {
	if x > len(r.text) {
		return
	}
	i, currentAttr := r.GetAttr(x)
	r.text = r.text[:x-1]
	currentAttr.end = x - 1
	r.attrs = r.attrs[:i+1]
}

func (r *Row) RemoveN(x, N int) {
	if x < 1 || N < 1 || x > len(r.text) {
		return
	}
	end := x + N - 1
	if end > len(r.text) {
		end = len(r.text)
	}
	a, attrA := r.GetAttr(x)
	b, attrB := r.GetAttr(end)
	r.text = append(r.text[:x-1], r.text[end:]...)

	for j := range r.attrs[b+1:] {
		r.attrs[b+1+j].start -= N
		r.attrs[b+1+j].end -= N
	}

	if a != b {
		removeA := a + 1
		removeB := b
		if attrA.start < x {
			attrA.end = x - 1
		} else {
			removeA = a
		}
		if attrB.end > end {
			attrB.start = x
			attrB.end -= N
		} else {
			removeB = b + 1
		}
		if removeA > 0 && removeB < len(r.attrs) && r.attrs[removeA-1].style == r.attrs[removeB].style {
			r.attrs[removeA-1].end = r.attrs[removeB].end
			removeB++
		}
		r.attrs = append(r.attrs[:removeA], r.attrs[removeB:]...)

	} else {
		if attrA.start == x && attrA.end == end {
			if a > 0 && a < len(r.attrs) && r.attrs[a-1].style == r.attrs[a+1].style {
				r.attrs[a-1].end = r.attrs[a+1].end
				r.attrs = append(r.attrs[:a], r.attrs[a+2:]...)
			} else {
				r.attrs = append(r.attrs[:a], r.attrs[a+1:]...)
			}
		} else {
			attrA.end -= N
		}
	}
}

/*
x=7, N = 5
len = 9
9-7 =
xyzxyzxyz

	      ^
		  12345
*/
func (r *Row) EraseToNO(x, N int) {
	length := len(r.text)
	if x > length {
		r.text = append(r.text, []rune(strings.Repeat(" ", x+N-length-1))...)
		r.attrs = append(r.attrs, Attr{start: length, end: x + N - 1, style: NewStyle()})
	} else if x+N-1 > length {
		_, currentAttr := r.GetAttr(x)
		for i := range length - x + 1 {
			r.text[x+i-1] = ' '
		}
		for range x + N - length - 1 {
			r.text = append(r.text, ' ')
		}
		if currentAttr != nil {
			currentAttr.end = x + N - 1
		}
	} else {
		for i := range N {
			r.text[x+i-1] = ' '
		}
	}
}

func (r *Row) EraseToN(x, N int) {
	style := NewStyle()
	for i := range N {
		r.AddText(' ', x+i, &style)
	}
}

func (r *Row) Html() string {
	var sb strings.Builder

	for _, attr := range r.attrs {
		if !attr.style.IsEmpty() {
			sb.WriteString("<span ")
			sb.WriteString(attr.style.Attributes())
			sb.WriteString(">")
		}
		sb.WriteString(string(r.text[attr.start-1 : attr.end]))
		if !attr.style.IsEmpty() {
			sb.WriteString("</span>")
		}
	}
	return sb.String()
}

func (r *Row) HtmlWithCursor(x int) string {
	var sb strings.Builder
	offset := 0
	for _, attr := range r.attrs {
		if !attr.style.IsEmpty() {
			sb.WriteString("<span class=\"")
			sb.WriteString(attr.style.Classes())
			sb.WriteString("\">")
		}
		if x >= attr.start && x <= attr.end {
			sb.WriteString(string(r.text[attr.start-1 : x-1]))
			sb.WriteString("<span class=\"cursor\">")
			sb.WriteRune(r.text[x-1])
			sb.WriteString("</span>")
			sb.WriteString(string(r.text[x:attr.end]))
		} else {
			sb.WriteString(string(r.text[attr.start-1 : attr.end]))
		}
		if !attr.style.IsEmpty() {
			sb.WriteString("</span>")
		}
		offset = attr.end
	}
	if x > offset {
		sb.WriteString(strings.Repeat(" ", x-offset-1))
		sb.WriteString("<span class=\"cursor\"> </span>")
	}

	return sb.String()
}
