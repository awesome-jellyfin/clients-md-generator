package generator

import (
	"fmt"
	"strings"
)

type MarkdownRenderer interface {
	Render() string
}

// Link is a markdown link.
type Link struct {
	Text MarkdownRenderer
	URL  string
}

func (l Link) Render() string {
	return fmt.Sprintf("[%s](%s)", l.Text.Render(), l.URL)
}

// Image is a markdown image.
type Image struct {
	AltText  MarkdownRenderer
	ImageURL string
}

func (i Image) Render() string {
	return fmt.Sprintf("![%s](%s)", i.AltText.Render(), i.ImageURL)
}

// Text is a markdown text.
// Options can be used to apply formatting to the text.
type Text struct {
	string

	Bold          bool
	Italic        bool
	Strikethrough bool
}

func (t Text) Render() string {
	var outer string
	if t.Bold {
		outer += "**"
	}
	if t.Italic {
		outer += "*"
	}
	if t.Strikethrough {
		outer += "~~"
	}
	return fmt.Sprintf("%s%s%s", outer, t.string, outer)
}

type TextOption func(*Text)

//goland:noinspection GoUnusedGlobalVariable
var (
	Bold = func(t *Text) {
		t.Bold = true
	}
	Italic = func(t *Text) {
		t.Italic = true
	}
	Strikethrough = func(t *Text) {
		t.Strikethrough = true
	}
)

func NewText(s string, options ...TextOption) Text {
	t := Text{string: s}
	for _, option := range options {
		option(&t)
	}
	return t
}

// Vertical is a vertical list of strings.
type Vertical []MarkdownRenderer

func (v *Vertical) Append(values ...MarkdownRenderer) {
	*v = append(*v, values...)
}

func (v *Vertical) Render() string {
	var bob strings.Builder
	for i, item := range *v {
		if i > 0 {
			bob.WriteString("\n\n")
		}
		bob.WriteString(item.Render())
	}
	return bob.String()
}

func NewVertical(values ...MarkdownRenderer) Vertical {
	return values
}

// Horizontal is a horizontal list of strings.
type Horizontal struct {
	Items []MarkdownRenderer
}

func (h *Horizontal) Append(values ...MarkdownRenderer) {
	h.Items = append(h.Items, values...)
}

func (h *Horizontal) Render() string {
	var bob strings.Builder
	for i, item := range h.Items {
		if i > 0 {
			bob.WriteString(" ")
		}
		bob.WriteString(item.Render())
	}
	return bob.String()
}

func NewHorizontal(values ...MarkdownRenderer) *Horizontal {
	return &Horizontal{Items: values}
}

// List is a markdown list.
// If Enumerated is true, the list will be enumerated.
type List struct {
	Enumerated bool
	Items      []MarkdownRenderer
}

func (l *List) Append(values ...MarkdownRenderer) {
	l.Items = append(l.Items, values...)
}

func (l *List) Render() string {
	var bob strings.Builder
	for i, item := range l.Items {
		if l.Enumerated {
			bob.WriteString(fmt.Sprintf("%d. ", i+1))
		} else {
			bob.WriteString("* ")
		}
		bob.WriteString(item.Render())
		bob.WriteString("\n")
	}
	return bob.String()
}

func NewList(enumerated bool) *List {
	return &List{Enumerated: enumerated}
}

// Heading is a markdown heading.
type Heading struct {
	Level int
	Text  MarkdownRenderer
}

func (h Heading) Render() string {
	return fmt.Sprintf("%s %s", strings.Repeat("#", h.Level), h.Text.Render())
}

func NewHeading(level int, text MarkdownRenderer) Heading {
	return Heading{Level: level, Text: text}
}

// Table is a markdown table.
type Table struct {
	Header []MarkdownRenderer
	Rows   [][]MarkdownRenderer
}

func (t Table) writeHeader(bob *strings.Builder) {
	var dividor strings.Builder
	dividor.WriteString("| ")

	bob.WriteString("| ")
	for i, header := range t.Header {
		if i > 0 {
			bob.WriteString(" | ")
			dividor.WriteString(" | ")
		}
		val := header.Render()
		bob.WriteString(val)

		l := len(val)
		if l < 3 {
			l = 3
		}
		dividor.WriteString(strings.Repeat("-", l))
	}
	bob.WriteString(" |\n")
	dividor.WriteString(" |\n")

	bob.WriteString(dividor.String())
}

func (t Table) Render() string {
	var bob strings.Builder

	t.writeHeader(&bob)
	for _, row := range t.Rows {
		bob.WriteString("| ")

		for i, cell := range row {
			if i > 0 {
				bob.WriteString(" | ")
			}
			bob.WriteString(cell.Render())
		}

		bob.WriteString(" |\n")
	}

	return bob.String()
}

type TableBuilder struct {
	Header []MarkdownRenderer
	Rows   [][]MarkdownRenderer
}

func NewTableBuilder(headers ...MarkdownRenderer) *TableBuilder {
	return &TableBuilder{Header: headers}
}

func (t *TableBuilder) AddRow(row ...MarkdownRenderer) {
	t.Rows = append(t.Rows, row)
}

func (t *TableBuilder) Build() Table {
	return Table{Header: t.Header, Rows: t.Rows}
}

type HorizontalDivider struct{}

func (h HorizontalDivider) Render() string {
	return "---"
}

type Code struct {
	string
	Padded bool
	Block  bool
}

func (c Code) Render() string {
	if c.Block {
		return fmt.Sprintf("```\n%s\n```", c.string)
	}
	if c.Padded {
		return fmt.Sprintf("` %s `", c.string)
	}
	return fmt.Sprintf("`%s`", c.string)
}

type CodeOption func(*Code)

//goland:noinspection GoUnusedGlobalVariable
var (
	CodePadded = func(c *Code) {
		c.Padded = true
	}
	CodeBlock = func(c *Code) {
		c.Block = true
	}
)

func NewCode(s string, options ...CodeOption) Code {
	c := Code{string: s}
	for _, option := range options {
		option(&c)
	}
	return c
}
